package contact

import (
	"context"
	"errors"
	"sort"

	"github.com/241x/zero-kit/apperror"
	"github.com/241x/zero-kit/baserepo"
	"github.com/241x/zero-web/errcode"
)

// Service 通讯录服务
type Service struct {
	departmentRepo       *WecomDepartmentRepository
	memberRepo           *WecomMemberRepository
	memberDepartmentRepo *WecomMemberDepartmentRepository
}

// NewService 创建通讯录服务
func NewService(departmentRepo *WecomDepartmentRepository, memberRepo *WecomMemberRepository, memberDepartmentRepo *WecomMemberDepartmentRepository) *Service {
	return &Service{departmentRepo: departmentRepo, memberRepo: memberRepo, memberDepartmentRepo: memberDepartmentRepo}
}

// GetDepartmentTree 获取部门树，附带各部门成员数
func (s *Service) GetDepartmentTree(ctx context.Context, storeId uint32) (*DepartmentTreeResponse, error) {
	departments, err := s.departmentRepo.FindAll(ctx, &DepartmentFilter{StoreId: storeId}, nil, nil)
	if err != nil {
		return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("获取部门树失败"))
	}

	// 统计各部门成员数（SQL层聚合）
	memberCountMap, err := s.memberDepartmentRepo.CountByDepartment(ctx, storeId)
	if err != nil {
		return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("统计部门成员数失败"))
	}

	// 内存建树
	nodeMap := make(map[uint32]*DepartmentNode, len(departments))
	for _, department := range departments {
		nodeMap[department.DepartmentId] = &DepartmentNode{
			ID:           department.ID,
			DepartmentId: department.DepartmentId,
			ParentId:     department.ParentId,
			Name:         department.Name,
			Sort:         department.Sort,
			MemberCount:  memberCountMap[department.DepartmentId],
			Children:     []*DepartmentNode{},
		}
	}

	roots := make([]*DepartmentNode, 0, len(departments))
	for _, node := range nodeMap {
		if parent, ok := nodeMap[node.ParentId]; ok {
			parent.Children = append(parent.Children, node)
		} else {
			roots = append(roots, node)
		}
	}

	sortNodes(roots)
	return &DepartmentTreeResponse{Tree: roots}, nil
}

// FindMemberList 成员分页列表
func (s *Service) FindMemberList(ctx context.Context, storeId uint32, req *MemberListRequest) (*MemberListResponse, error) {
	result := &MemberListResponse{List: []*MemberInfo{}, Total: 0}

	var list []*WecomMember
	var total int64
	var err error

	if req.DepartmentId != 0 {
		// 部门筛选：先取部门path，再以子查询分页查询成员（含子孙部门）
		department, err := s.departmentRepo.FindOne(ctx, &DepartmentFilter{StoreId: storeId, DepartmentId: req.DepartmentId})
		if err != nil {
			if errors.Is(err, baserepo.ErrRecordNotFound) {
				return result, nil
			}
			return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("获取成员列表失败"))
		}
		if department.Path == "" {
			return result, nil
		}
		list, total, err = s.memberRepo.FindPageByDepartmentPath(ctx, storeId, department.Path, req.Status, req.Page, req.Limit)
		if err != nil {
			return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("获取成员列表失败"))
		}
	} else {
		// 无部门筛选：常规分页
		filter := &MemberFilter{StoreId: storeId, Status: req.Status}
		total, err = s.memberRepo.Count(ctx, filter)
		if err != nil {
			return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("获取成员列表失败"))
		}
		if total == 0 {
			return result, nil
		}
		orders := baserepo.Orders{{Field: "id", Sort: "asc"}}
		list, err = s.memberRepo.FindAll(ctx, filter, baserepo.NewPagination(req.Page, req.Limit), orders)
		if err != nil {
			return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("获取成员列表失败"))
		}
	}
	if total == 0 || len(list) == 0 {
		return result, nil
	}

	// 批量查询部门名映射
	departmentNameMap, err := s.buildDepartmentNameMap(ctx, storeId, nil)
	if err != nil {
		return nil, err
	}

	// 仅查询当前页成员的部门关联，避免全量拉取
	currentUserIds := make([]string, 0, len(list))
	for _, member := range list {
		currentUserIds = append(currentUserIds, member.UserId)
	}
	pageRelations, err := s.memberDepartmentRepo.FindListByUserIds(ctx, storeId, currentUserIds)
	if err != nil {
		return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("获取成员列表失败"))
	}
	relationByUser := make(map[string][]*WecomMemberDepartment, len(pageRelations))
	for _, rel := range pageRelations {
		relationByUser[rel.UserId] = append(relationByUser[rel.UserId], rel)
	}

	result.Total = total
	result.List = make([]*MemberInfo, 0, len(list))
	for _, member := range list {
		info := &MemberInfo{
			UserId:   member.UserId,
			Name:     member.Name,
			Position: member.Position,
			Avatar:   member.Avatar,
			Status:   member.Status,
		}
		// 补部门信息
		names := make([]string, 0, 2)
		for _, rel := range relationByUser[member.UserId] {
			if name, ok := departmentNameMap[rel.DepartmentId]; ok && name != "" {
				names = append(names, name)
			}
			if rel.IsLeader == 1 {
				info.IsLeader = true
			}
		}
		info.Departments = names
		result.List = append(result.List, info)
	}
	return result, nil
}

// buildDepartmentNameMap 构建部门ID->名称映射。departmentIds 为空时查询全部部门
func (s *Service) buildDepartmentNameMap(ctx context.Context, storeId uint32, departmentIds []uint32) (map[uint32]string, error) {
	filter := &DepartmentFilter{StoreId: storeId}
	if len(departmentIds) > 0 {
		filter.DepartmentIds = departmentIds
	}
	departments, err := s.departmentRepo.FindAll(ctx, filter, nil, nil)
	if err != nil {
		return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("获取部门信息失败"))
	}
	nameMap := make(map[uint32]string, len(departments))
	for _, department := range departments {
		nameMap[department.DepartmentId] = department.Name
	}
	return nameMap, nil
}

// FindMemberDetail 成员详情（含所属部门）
func (s *Service) FindMemberDetail(ctx context.Context, storeId uint32, req *MemberDetailRequest) (*MemberDetailResponse, error) {
	member, err := s.memberRepo.FindOne(ctx, &MemberFilter{StoreId: storeId, UserId: req.UserId})
	if err != nil {
		if errors.Is(err, baserepo.ErrRecordNotFound) {
			return nil, apperror.New(errcode.NotFound, apperror.WithMsg("成员不存在"))
		}
		return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("获取成员详情失败"))
	}

	// 查询所属部门
	relations, err := s.memberDepartmentRepo.FindAll(ctx, &MemberDepartmentFilter{StoreId: storeId, UserId: req.UserId}, nil, nil)
	if err != nil {
		return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("获取成员详情失败"))
	}

	// 只查询该成员涉及的部门，避免全量拉取
	departmentIds := make([]uint32, 0, len(relations))
	for _, rel := range relations {
		departmentIds = append(departmentIds, rel.DepartmentId)
	}
	departmentNameMap := map[uint32]string{}
	if len(departmentIds) > 0 {
		departmentNameMap, err = s.buildDepartmentNameMap(ctx, storeId, departmentIds)
		if err != nil {
			return nil, err
		}
	}

	departments := make([]*MemberDepartmentInfo, 0, len(relations))
	for _, rel := range relations {
		departments = append(departments, &MemberDepartmentInfo{
			DepartmentId: rel.DepartmentId,
			Name:         departmentNameMap[rel.DepartmentId],
			IsLeader:     rel.IsLeader == 1,
		})
	}

	return &MemberDetailResponse{
		UserId:      member.UserId,
		Name:        member.Name,
		Position:    member.Position,
		Mobile:      member.Mobile,
		Gender:      member.Gender,
		Email:       member.Email,
		Avatar:      member.Avatar,
		Status:      member.Status,
		Departments: departments,
	}, nil
}

// sortNodes 按 Sort 排序部门节点
func sortNodes(nodes []*DepartmentNode) {
	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].Sort != nodes[j].Sort {
			return nodes[i].Sort < nodes[j].Sort
		}
		return nodes[i].DepartmentId < nodes[j].DepartmentId
	})
	for _, node := range nodes {
		sortNodes(node.Children)
	}
}
