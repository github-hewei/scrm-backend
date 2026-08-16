package group

import (
	"context"
	"fmt"

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

// Sync 全量同步客户群。策略：分页拉取群列表，逐个拉详情并 upsert 群与成员；
// 全部分页拉取成功后，清理企微已不返回（解散）的本地群及其成员
func (s *GroupSyncer) Sync(ctx context.Context, client *wecom.Client, storeId uint32) error {
	const maxPages = 100000 // 防御：游标异常时避免死循环（10万页×100≈千万群，正常场景远达不到）
	cursor := ""
	seen := make(map[string]struct{})
	page := 0
	for ; page < maxPages; page++ {
		listResp, err := client.ExternalContact.GetGroupChatList(ctx, 0, 100, nil, cursor)
		if err != nil {
			return apperror.Wrap(errcode.Internal, err, apperror.WithMsgf("拉取客户群列表失败 store_id=%d", storeId))
		}
		for _, entry := range listResp.GroupChatList {
			seen[entry.ChatID] = struct{}{}
			if err := s.syncGroup(ctx, client, storeId, entry.ChatID, int8(entry.Status)); err != nil {
				return err
			}
		}
		if listResp.NextCursor == "" {
			break
		}
		cursor = listResp.NextCursor
	}
	if page == maxPages {
		return apperror.Wrap(errcode.Internal, fmt.Errorf("拉取客户群列表分页超限 store_id=%d", storeId))
	}
	if len(seen) == 0 {
		// 拉取源为空（企微异常/未返回任何群）时不执行清理，避免误删本地已有客户群
		return nil
	}
	return s.cleanupDisbanded(ctx, storeId, seen)
}

// cleanupDisbanded 清理已解散群：本地存在但企微列表不再返回的群软删，成员保留（软删群对程序不可见，成员无独立访问路径）。
// 仅在整轮分页拉取成功后执行，避免中途失败时误删未同步到的群
func (s *GroupSyncer) cleanupDisbanded(ctx context.Context, storeId uint32, seen map[string]struct{}) error {
	groups, err := s.groupRepo.FindAll(ctx, &GroupFilter{StoreId: storeId}, nil, nil)
	if err != nil {
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsgf("查询本地客户群失败 store_id=%d", storeId))
	}
	for _, g := range groups {
		if _, ok := seen[g.ChatId]; ok {
			continue
		}
		if err := s.groupRepo.Delete(ctx, g.ID); err != nil {
			return apperror.Wrap(errcode.Internal, err, apperror.WithMsgf("清理已解散客户群失败 chat_id=%s", g.ChatId))
		}
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
// 批量更新已存在、批量插入新增、批量软删除已移除。
// 软删行对程序不可见，退群后重新入群的成员视为新成员直接插入
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
