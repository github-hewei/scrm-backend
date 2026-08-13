package group

import (
	"context"

	"github.com/241x/zero-kit/apperror"
	"github.com/241x/zero-kit/baserepo"
	"github.com/241x/zero-third/wecom"
	"github.com/241x/zero-web/errcode"
)

// GroupSyncer 客户群同步器
type GroupSyncer struct {
	groupRepo       *WecomGroupRepository
	groupMemberRepo *WecomGroupMemberRepository
}

// NewGroupSyncer 创建客户群同步器
func NewGroupSyncer(groupRepo *WecomGroupRepository, groupMemberRepo *WecomGroupMemberRepository) *GroupSyncer {
	return &GroupSyncer{groupRepo: groupRepo, groupMemberRepo: groupMemberRepo}
}

// Sync 全量同步客户群。策略：分页拉取群列表，逐个拉详情并 upsert 群与成员
func (s *GroupSyncer) Sync(ctx context.Context, client *wecom.Client, storeId uint32) error {
	cursor := ""
	for {
		listResp, err := client.ExternalContact.GetGroupChatList(ctx, 0, 100, nil, cursor)
		if err != nil {
			return apperror.Wrap(errcode.Internal, err, apperror.WithMsgf("拉取客户群列表失败 store_id=%d", storeId))
		}
		for _, entry := range listResp.GroupChatList {
			if err := s.syncGroup(ctx, client, storeId, entry.ChatID, int8(entry.Status)); err != nil {
				return err
			}
		}
		if listResp.NextCursor == "" {
			break
		}
		cursor = listResp.NextCursor
	}
	return nil
}

// syncGroup 同步单个客户群及其成员
func (s *GroupSyncer) syncGroup(ctx context.Context, client *wecom.Client, storeId uint32, chatId string, status int8) error {
	detail, err := client.ExternalContact.GetGroupChatDetail(ctx, chatId, true)
	if err != nil {
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsgf("拉取客户群详情失败 chat_id=%s", chatId))
	}

	// upsert 群
	if err := s.upsertGroup(ctx, storeId, chatId, status, detail); err != nil {
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsgf("写入客户群失败 chat_id=%s", chatId))
	}

	// upsert 成员（增量，复用主键ID）
	if err := s.upsertMembers(ctx, storeId, chatId, detail); err != nil {
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsgf("写入群成员失败 chat_id=%s", chatId))
	}
	return nil
}

// upsertGroup 按 chat_id upsert 客户群
func (s *GroupSyncer) upsertGroup(ctx context.Context, storeId uint32, chatId string, status int8, detail *wecom.GroupChat) error {
	existing, err := s.groupRepo.FindOne(ctx, &GroupFilter{StoreId: storeId, ChatId: chatId})
	if err != nil && err != baserepo.ErrRecordNotFound {
		return err
	}
	updateData := map[string]any{
		"name": detail.Name, "owner": detail.Owner, "create_time": uint32(detail.CreateTime),
		"notice": detail.Notice, "status": status, "member_count": uint32(len(detail.MemberList)),
		"member_version": detail.MemberVersion,
	}
	if existing != nil {
		return s.groupRepo.Updates(ctx, existing, updateData)
	}
	return s.groupRepo.Create(ctx, &WecomGroup{
		StoreId: storeId, ChatId: chatId, Name: detail.Name, Owner: detail.Owner,
		CreateTime: uint32(detail.CreateTime), Notice: detail.Notice, Status: status,
		MemberCount: uint32(len(detail.MemberList)), MemberVersion: detail.MemberVersion,
	})
}

// upsertMembers 按 (store_id, chat_id, user_id) 增量同步群成员：
// 批量更新已存在、批量插入新增、批量软删除已移除
func (s *GroupSyncer) upsertMembers(ctx context.Context, storeId uint32, chatId string, detail *wecom.GroupChat) error {
	// 现有成员映射
	existingList, err := s.groupMemberRepo.FindAll(ctx, &GroupMemberFilter{StoreId: storeId, ChatId: chatId}, nil, nil)
	if err != nil {
		return err
	}
	existing := make(map[string]*WecomGroupMember, len(existingList))
	for _, member := range existingList {
		existing[member.UserId] = member
	}

	// 管理员集合（map化，O(1)判断）
	adminSet := make(map[string]struct{}, len(detail.AdminList))
	for _, admin := range detail.AdminList {
		adminSet[admin.UserID] = struct{}{}
	}

	// 分批差异：更新/新增/删除
	toUpdate := make([]*WecomGroupMember, 0, len(detail.MemberList))
	toCreate := make([]*WecomGroupMember, 0, len(detail.MemberList))
	toDelete := make([]*WecomGroupMember, 0)
	incoming := make(map[string]struct{}, len(detail.MemberList))

	for _, m := range detail.MemberList {
		incoming[m.UserID] = struct{}{}
		model := &WecomGroupMember{
			StoreId: storeId, ChatId: chatId, UserId: m.UserID,
			Type: int8(m.Type), Unionid: m.UnionID, JoinTime: uint32(m.JoinTime),
			JoinScene: int8(m.JoinScene), InvitorUserid: invitorOf(m.Invitor),
			GroupNickname: m.GroupNickname, Name: m.Name,
			IsAdmin: boolToInt8(containsUserID(adminSet, m.UserID)),
		}
		if old, ok := existing[m.UserID]; ok {
			model.ID = old.ID // 保留主键，批量更新用
			toUpdate = append(toUpdate, model)
		} else {
			toCreate = append(toCreate, model)
		}
	}
	for userId, old := range existing {
		if _, ok := incoming[userId]; !ok {
			toDelete = append(toDelete, old)
		}
	}

	// 批量写
	if len(toUpdate) > 0 {
		for _, member := range toUpdate {
			if err := s.groupMemberRepo.Updates(ctx, member, map[string]any{
				"type": member.Type, "unionid": member.Unionid, "join_time": member.JoinTime,
				"join_scene": member.JoinScene, "invitor_userid": member.InvitorUserid,
				"group_nickname": member.GroupNickname, "name": member.Name, "is_admin": member.IsAdmin,
			}); err != nil {
				return err
			}
		}
	}
	if len(toCreate) > 0 {
		if err := s.groupMemberRepo.CreateBatch(ctx, toCreate); err != nil {
			return err
		}
	}
	if len(toDelete) > 0 {
		for _, member := range toDelete {
			if err := s.groupMemberRepo.Delete(ctx, member.ID); err != nil {
				return err
			}
		}
	}
	return nil
}

// invitorOf 提取邀请人UserID，兼容空指针
func invitorOf(invitor *wecom.GroupChatInvitor) string {
	if invitor == nil {
		return ""
	}
	return invitor.UserID
}

// containsUserID 判断userID是否在管理员集合中（map化后O(1)）
func containsUserID(adminSet map[string]struct{}, userID string) bool {
	_, ok := adminSet[userID]
	return ok
}

// boolToInt8 布尔转0/1
func boolToInt8(b bool) int8 {
	if b {
		return 1
	}
	return 0
}
