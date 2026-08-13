package group

import (
	"context"
	"errors"

	"github.com/241x/zero-kit/apperror"
	"github.com/241x/zero-kit/baserepo"
	"github.com/241x/zero-web/errcode"
)

// Service 客户群服务
type Service struct {
	groupRepo       *WecomGroupRepository
	groupMemberRepo *WecomGroupMemberRepository
}

// NewService 创建客户群服务
func NewService(groupRepo *WecomGroupRepository, groupMemberRepo *WecomGroupMemberRepository) *Service {
	return &Service{groupRepo: groupRepo, groupMemberRepo: groupMemberRepo}
}

// FindList 客户群分页列表
func (s *Service) FindList(ctx context.Context, storeId uint32, req *GroupListRequest) (*GroupListResponse, error) {
	result := &GroupListResponse{List: []*GroupInfo{}, Total: 0}

	// 列表过滤：status=-1 表示全部（由GroupListFilter内部处理）
	filter := &GroupListFilter{StoreId: storeId, Owner: req.Owner, Status: req.Status}
	total, err := s.groupRepo.Count(ctx, filter)
	if err != nil {
		return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("获取客户群列表失败"))
	}
	if total == 0 {
		return result, nil
	}

	orders := baserepo.Orders{{Field: "id", Sort: "desc"}}
	list, err := s.groupRepo.FindAll(ctx, filter, baserepo.NewPagination(req.Page, req.Limit), orders)
	if err != nil {
		return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("获取客户群列表失败"))
	}

	// 批量查询群主姓名（群主必为群内企业成员）
	ownerIds := make([]string, 0, len(list))
	for _, group := range list {
		if group.Owner != "" {
			ownerIds = append(ownerIds, group.Owner)
		}
	}
	ownerNameMap, err := s.groupMemberRepo.FindMemberNames(ctx, storeId, ownerIds)
	if err != nil {
		return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("获取群主信息失败"))
	}

	result.Total = total
	result.List = make([]*GroupInfo, 0, len(list))
	for _, group := range list {
		result.List = append(result.List, &GroupInfo{
			ChatId:      group.ChatId,
			Name:        group.Name,
			Owner:       group.Owner,
			OwnerName:   ownerNameMap[group.Owner],
			MemberCount: group.MemberCount,
			Status:      group.Status,
			StatusText:  GroupStatus(group.Status).Name(),
			CreateTime:  group.CreateTime,
		})
	}
	return result, nil
}

// FindDetail 客户群详情（含成员列表）
func (s *Service) FindDetail(ctx context.Context, storeId uint32, req *GroupDetailRequest) (*GroupDetailResponse, error) {
	group, err := s.groupRepo.FindOne(ctx, &GroupFilter{StoreId: storeId, ChatId: req.ChatId})
	if err != nil {
		if errors.Is(err, baserepo.ErrRecordNotFound) {
			return nil, apperror.New(errcode.NotFound, apperror.WithMsg("客户群不存在"))
		}
		return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("获取客户群详情失败"))
	}

	// 查询群主姓名
	ownerNameMap, err := s.groupMemberRepo.FindMemberNames(ctx, storeId, []string{group.Owner})
	if err != nil {
		return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("获取群主信息失败"))
	}

	// 查询全部群成员（量级可控，不分页）
	members, err := s.groupMemberRepo.FindAll(ctx, &GroupMemberFilter{StoreId: storeId, ChatId: req.ChatId}, nil, nil)
	if err != nil {
		return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("获取客户群详情失败"))
	}

	memberInfos := make([]*GroupMemberInfo, 0, len(members))
	for _, member := range members {
		memberInfos = append(memberInfos, &GroupMemberInfo{
			UserId:        member.UserId,
			Name:          member.Name,
			Type:          member.Type,
			TypeText:      GroupMemberType(member.Type).Name(),
			JoinTime:      member.JoinTime,
			JoinScene:     member.JoinScene,
			JoinSceneText: JoinScene(member.JoinScene).Name(),
			IsAdmin:       member.IsAdmin == 1,
		})
	}

	return &GroupDetailResponse{
		ChatId:      group.ChatId,
		Name:        group.Name,
		Owner:       group.Owner,
		OwnerName:   ownerNameMap[group.Owner],
		Notice:      group.Notice,
		Status:      group.Status,
		StatusText:  GroupStatus(group.Status).Name(),
		MemberCount: group.MemberCount,
		CreateTime:  group.CreateTime,
		Members:     memberInfos,
	}, nil
}
