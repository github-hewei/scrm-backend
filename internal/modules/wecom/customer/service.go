package customer

import (
	"context"
	"errors"

	"github.com/241x/zero-kit/apperror"
	"github.com/241x/zero-kit/baserepo"
	"github.com/241x/zero-web/errcode"
)

// MemberNameProvider 成员姓名查询接口，由外部注入实现（跨包查询通讯录成员名）
type MemberNameProvider interface {
	FindMemberNames(ctx context.Context, storeId uint32, userIds []string) (map[string]string, error)
}

// Service 客户服务
type Service struct {
	customerRepo *WecomCustomerRepository
	followRepo   *WecomCustomerFollowRepository
	memberNames  MemberNameProvider
}

// NewService 创建客户服务
func NewService(customerRepo *WecomCustomerRepository, followRepo *WecomCustomerFollowRepository, memberNames MemberNameProvider) *Service {
	return &Service{customerRepo: customerRepo, followRepo: followRepo, memberNames: memberNames}
}

// FindList 客户分页列表（纯客户主数据）
func (s *Service) FindList(ctx context.Context, storeId uint32, req *CustomerListRequest) (*CustomerListResponse, error) {
	result := &CustomerListResponse{List: []*CustomerInfo{}, Total: 0}
	filter := &CustomerFilter{StoreId: storeId, Type: req.Type}
	total, err := s.customerRepo.Count(ctx, filter)
	if err != nil {
		return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("获取客户列表失败"))
	}
	if total == 0 {
		return result, nil
	}
	orders := baserepo.Orders{{Field: "id", Sort: "desc"}}
	list, err := s.customerRepo.FindAll(ctx, filter, baserepo.NewPagination(req.Page, req.Limit), orders)
	if err != nil {
		return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("获取客户列表失败"))
	}
	result.Total = total
	result.List = make([]*CustomerInfo, 0, len(list))
	for _, customer := range list {
		result.List = append(result.List, &CustomerInfo{
			ExternalUserid: customer.ExternalUserid,
			Name:           customer.Name,
			Avatar:         customer.Avatar,
			Type:           customer.Type,
			Gender:         customer.Gender,
			GenderText:     Gender(customer.Gender).Name(),
			CorpName:       customer.CorpName,
		})
	}
	return result, nil
}

// FindDetail 客户详情（含跟进人）
func (s *Service) FindDetail(ctx context.Context, storeId uint32, req *CustomerDetailRequest) (*CustomerDetailResponse, error) {
	customer, err := s.customerRepo.FindOne(ctx, &CustomerFilter{StoreId: storeId, ExternalUserid: req.ExternalUserid})
	if err != nil {
		if errors.Is(err, baserepo.ErrRecordNotFound) {
			return nil, apperror.New(errcode.NotFound, apperror.WithMsg("客户不存在"))
		}
		return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("获取客户详情失败"))
	}

	// 查询跟进人
	follows, err := s.followRepo.FindAll(ctx, &CustomerFollowFilter{StoreId: storeId, ExternalUserid: req.ExternalUserid}, nil, nil)
	if err != nil {
		return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("获取客户详情失败"))
	}

	// 批量查询跟进人姓名
	userIds := make([]string, 0, len(follows))
	for _, follow := range follows {
		userIds = append(userIds, follow.UserId)
	}
	nameMap := map[string]string{}
	if s.memberNames != nil && len(userIds) > 0 {
		nameMap, err = s.memberNames.FindMemberNames(ctx, storeId, userIds)
		if err != nil {
			return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("获取跟进人信息失败"))
		}
	}

	followInfos := make([]*FollowInfo, 0, len(follows))
	for _, follow := range follows {
		followInfos = append(followInfos, &FollowInfo{
			UserId:         follow.UserId,
			UserName:       nameMap[follow.UserId],
			Remark:         follow.Remark,
			Description:    follow.Description,
			RemarkCorpName: follow.RemarkCorpName,
			RemarkMobiles:  follow.RemarkMobiles,
			AddWay:         follow.AddWay,
			AddWayText:     AddWay(follow.AddWay).Name(),
			State:          follow.State,
			CreateTime:     follow.CreateTime,
		})
	}

	return &CustomerDetailResponse{
		ExternalUserid: customer.ExternalUserid,
		Name:           customer.Name,
		Avatar:         customer.Avatar,
		Type:           customer.Type,
		Gender:         customer.Gender,
		GenderText:     Gender(customer.Gender).Name(),
		Position:       customer.Position,
		CorpName:       customer.CorpName,
		CorpFullName:   customer.CorpFullName,
		Unionid:        customer.Unionid,
		Follows:        followInfos,
	}, nil
}
