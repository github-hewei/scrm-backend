package customer

import (
	"context"
	"fmt"
	"strings"

	"github.com/241x/zero-kit/apperror"
	"github.com/241x/zero-third/wecom"
	"github.com/241x/zero-web/errcode"
)

// 批量拉取参数：batch/get_by_user 入参 userid_list 上限 100，单页返回上限 1000
const (
	followUserBatchSize = 100
	batchLimit          = 1000
	// maxBatchPages 单批翻页上限，防御游标异常死循环（1000页×1000≈100万客户，正常场景远达不到）
	maxBatchPages = 1000
)

// customerAggregate 客户聚合结果：主数据 + 全部跟进人（多成员视角合并）
type customerAggregate struct {
	contact *wecom.ExternalContact
	follows []wecom.FollowUser
	// followsByUser 跟进人按成员去重（防御企微返回重复）
	followsByUser map[string]struct{}
}

// CustomerSyncer 企业微信客户同步器
type CustomerSyncer struct {
	customerRepo *WecomCustomerRepository
	followRepo   *WecomCustomerFollowRepository
}

// NewCustomerSyncer 创建客户同步器
func NewCustomerSyncer(customerRepo *WecomCustomerRepository, followRepo *WecomCustomerFollowRepository) *CustomerSyncer {
	return &CustomerSyncer{customerRepo: customerRepo, followRepo: followRepo}
}

// Sync 全量同步客户。策略：
// 阶段一：拉取客户联系成员，分批(100/批)调批量详情接口翻页聚合客户与跟进人；
// 阶段二：一次读现有数据做差异批量写（无变化跳过，避免全量更新风暴）；
// 全部成功后清理企微已不返回（被所有成员删除）的客户
func (s *CustomerSyncer) Sync(ctx context.Context, client *wecom.Client, storeId uint32) error {
	userIds, err := client.ExternalContact.GetFollowUserList(ctx)
	if err != nil {
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsgf("拉取客户联系成员失败 store_id=%d", storeId))
	}
	if len(userIds) == 0 {
		// 拉取源为空时直接返回，不做写入与清理：
		// 可能是企业未配置客户联系功能或接口权限异常，避免误删本地已有客户
		return nil
	}

	// 阶段一：分批拉取并聚合
	aggregated := make(map[string]*customerAggregate)
	for start := 0; start < len(userIds); start += followUserBatchSize {
		end := start + followUserBatchSize
		if end > len(userIds) {
			end = len(userIds)
		}
		if err := s.fetchBatch(ctx, client, userIds[start:end], aggregated); err != nil {
			return err
		}
	}

	// 阶段二：批量差异写
	if err := s.upsertAll(ctx, storeId, aggregated); err != nil {
		return err
	}
	return s.cleanupRemoved(ctx, storeId, aggregated)
}

// fetchBatch 批量拉取一批成员的全部客户信息并聚合（内部翻页）
func (s *CustomerSyncer) fetchBatch(ctx context.Context, client *wecom.Client, userIds []string, aggregated map[string]*customerAggregate) error {
	cursor := ""
	for page := 0; page < maxBatchPages; page++ {
		resp, err := client.ExternalContact.BatchGetContacts(ctx, userIds, cursor, batchLimit)
		if err != nil {
			return apperror.Wrap(errcode.Internal, err, apperror.WithMsgf("批量拉取客户失败 user_ids=%v", userIds))
		}
		for _, item := range resp.ExternalContactList {
			id := item.ExternalContact.ExternalUserID
			if id == "" {
				continue // 防御：无客户ID的记录不落库
			}
			agg := aggregated[id]
			if agg == nil {
				c := item.ExternalContact
				agg = &customerAggregate{contact: &c}
				aggregated[id] = agg
			}
			// 同一客户被多个成员跟进时，每个成员视角各一条跟进关系。
			// 批量接口的跟进信息不含标签明细（仅 tag_id），当前不落库标签
			info := item.FollowInfo
			if info.UserID == "" {
				continue // 防御：无跟进成员视角的记录不落库
			}
			if agg.followsByUser == nil {
				agg.followsByUser = make(map[string]struct{})
			}
			if _, dup := agg.followsByUser[info.UserID]; dup {
				continue // 防御：同一跟进关系重复出现时去重
			}
			agg.followsByUser[info.UserID] = struct{}{}
			agg.follows = append(agg.follows, wecom.FollowUser{
				UserID:         info.UserID,
				Remark:         info.Remark,
				Description:    info.Description,
				CreateTime:     info.CreateTime,
				RemarkCorpName: info.RemarkCorpName,
				RemarkMobiles:  info.RemarkMobiles,
				OperUserID:     info.OperUserID,
				AddWay:         info.AddWay,
				State:          info.State,
			})
		}
		if resp.NextCursor == "" {
			return nil
		}
		cursor = resp.NextCursor
	}
	return apperror.Wrap(errcode.Internal, fmt.Errorf("批量拉取客户分页超限 user_ids=%v", userIds))
}

// upsertAll 批量差异写：一次读现有客户/跟进人，分桶后批量插入、按主键更新、软删移除
func (s *CustomerSyncer) upsertAll(ctx context.Context, storeId uint32, aggregated map[string]*customerAggregate) error {
	// 一次读现有数据
	existingCustomers, err := s.customerRepo.FindAll(ctx, &CustomerFilter{StoreId: storeId}, nil, nil)
	if err != nil {
		return err
	}
	existingFollows, err := s.followRepo.FindAll(ctx, &CustomerFollowFilter{StoreId: storeId}, nil, nil)
	if err != nil {
		return err
	}

	customerMap := make(map[string]*WecomCustomer, len(existingCustomers))
	for _, c := range existingCustomers {
		customerMap[c.ExternalUserid] = c
	}
	followMap := make(map[string]map[string]*WecomCustomerFollow)
	for _, f := range existingFollows {
		if followMap[f.ExternalUserid] == nil {
			followMap[f.ExternalUserid] = make(map[string]*WecomCustomerFollow)
		}
		followMap[f.ExternalUserid][f.UserId] = f
	}

	// 客户差异分桶
	toCreateCustomers := make([]*WecomCustomer, 0, len(aggregated))
	toUpdateCustomers := make([]*WecomCustomer, 0)
	for id, agg := range aggregated {
		contact := agg.contact
		model := &WecomCustomer{
			StoreId: storeId, ExternalUserid: id,
			Name: contact.Name, Position: contact.Position, Avatar: contact.Avatar,
			CorpName: contact.CorpName, CorpFullName: contact.CorpFullName,
			Type: int8(contact.Type), Gender: int8(contact.Gender), Unionid: contact.UnionID,
		}
		if old, ok := customerMap[id]; ok {
			model.ID = old.ID
			if customerChanged(old, contact) {
				toUpdateCustomers = append(toUpdateCustomers, model)
			}
		} else {
			toCreateCustomers = append(toCreateCustomers, model)
		}
	}
	if len(toCreateCustomers) > 0 {
		if err := s.customerRepo.CreateBatch(ctx, toCreateCustomers); err != nil {
			return err
		}
	}
	for _, c := range toUpdateCustomers {
		if err := s.customerRepo.Updates(ctx, c, map[string]any{
			"name": c.Name, "position": c.Position, "avatar": c.Avatar,
			"corp_name": c.CorpName, "corp_full_name": c.CorpFullName,
			"type": c.Type, "gender": c.Gender, "unionid": c.Unionid,
		}); err != nil {
			return err
		}
	}

	// 跟进人差异分桶
	toCreateFollows := make([]*WecomCustomerFollow, 0)
	toUpdateFollows := make([]*WecomCustomerFollow, 0)
	toDeleteFollows := make([]*WecomCustomerFollow, 0)
	for id, agg := range aggregated {
		existing := followMap[id]
		incoming := make(map[string]struct{}, len(agg.follows))
		for _, f := range agg.follows {
			incoming[f.UserID] = struct{}{}
			model := &WecomCustomerFollow{
				StoreId: storeId, ExternalUserid: id, UserId: f.UserID,
				Remark: f.Remark, Description: f.Description, RemarkCorpName: f.RemarkCorpName,
				RemarkMobiles: strings.Join(f.RemarkMobiles, ","),
				AddWay:        int16(f.AddWay), State: f.State, CreateTime: uint32(f.CreateTime),
			}
			if old, ok := existing[f.UserID]; ok {
				model.ID = old.ID
				if followChanged(old, f) {
					toUpdateFollows = append(toUpdateFollows, model)
				}
			} else {
				toCreateFollows = append(toCreateFollows, model)
			}
		}
		for userId, old := range existing {
			if _, ok := incoming[userId]; !ok {
				toDeleteFollows = append(toDeleteFollows, old)
			}
		}
	}
	if len(toCreateFollows) > 0 {
		if err := s.followRepo.CreateBatch(ctx, toCreateFollows); err != nil {
			return err
		}
	}
	for _, f := range toUpdateFollows {
		if err := s.followRepo.Updates(ctx, f, map[string]any{
			"remark": f.Remark, "description": f.Description,
			"remark_corp_name": f.RemarkCorpName, "remark_mobiles": f.RemarkMobiles,
			"add_way": f.AddWay, "state": f.State, "create_time": f.CreateTime,
		}); err != nil {
			return err
		}
	}
	for _, f := range toDeleteFollows {
		if err := s.followRepo.Delete(ctx, f.ID); err != nil {
			return err
		}
	}
	return nil
}

// cleanupRemoved 清理已移除客户：本地存在但企微不再返回（被所有成员删除）的客户软删，
// 跟进人保留（软删客户对程序不可见，无独立访问路径）。仅在整轮拉取成功后执行
func (s *CustomerSyncer) cleanupRemoved(ctx context.Context, storeId uint32, aggregated map[string]*customerAggregate) error {
	list, err := s.customerRepo.FindAll(ctx, &CustomerFilter{StoreId: storeId}, nil, nil)
	if err != nil {
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsgf("查询本地客户失败 store_id=%d", storeId))
	}
	for _, c := range list {
		if _, ok := aggregated[c.ExternalUserid]; ok {
			continue
		}
		if err := s.customerRepo.Delete(ctx, c.ID); err != nil {
			return apperror.Wrap(errcode.Internal, err, apperror.WithMsgf("清理已移除客户失败 external_userid=%s", c.ExternalUserid))
		}
	}
	return nil
}

// customerChanged 客户主数据是否有变化（无变化跳过更新，避免全量更新风暴）
func customerChanged(old *WecomCustomer, c *wecom.ExternalContact) bool {
	return old.Name != c.Name ||
		old.Position != c.Position ||
		old.Avatar != c.Avatar ||
		old.CorpName != c.CorpName ||
		old.CorpFullName != c.CorpFullName ||
		old.Type != int8(c.Type) ||
		old.Gender != int8(c.Gender) ||
		old.Unionid != c.UnionID
}

// followChanged 跟进人是否有变化
func followChanged(old *WecomCustomerFollow, f wecom.FollowUser) bool {
	return old.Remark != f.Remark ||
		old.Description != f.Description ||
		old.RemarkCorpName != f.RemarkCorpName ||
		old.RemarkMobiles != strings.Join(f.RemarkMobiles, ",") ||
		old.AddWay != int16(f.AddWay) ||
		old.State != f.State ||
		old.CreateTime != uint32(f.CreateTime)
}
