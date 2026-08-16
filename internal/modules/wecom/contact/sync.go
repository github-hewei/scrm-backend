package contact

import (
	"context"
	"strconv"

	"github.com/241x/zero-kit/apperror"
	"github.com/241x/zero-third/wecom"
	"github.com/241x/zero-web/errcode"
)

// deptMemberRel 成员部门关联（拉取期中间结构）
type deptMemberRel struct {
	UserId       string
	DepartmentId uint32
	IsLeader     int8
	Sort         uint32
}

// ContactSyncer 企业微信通讯录同步器（部门 + 成员 + 成员部门关联）
type ContactSyncer struct {
	deptRepo       *WecomDepartmentRepository
	memberRepo     *WecomMemberRepository
	memberDeptRepo *WecomMemberDepartmentRepository
}

// NewContactSyncer 创建通讯录同步器
func NewContactSyncer(deptRepo *WecomDepartmentRepository, memberRepo *WecomMemberRepository, memberDeptRepo *WecomMemberDepartmentRepository) *ContactSyncer {
	return &ContactSyncer{deptRepo: deptRepo, memberRepo: memberRepo, memberDeptRepo: memberDeptRepo}
}

// Sync 全量同步通讯录。策略：
// 阶段一：拉取全部部门（含物化路径计算），逐部门分页拉直属成员（跨部门按 user_id 去重）并汇总部门关联；
// 阶段二：一次读现有数据做差异批量写；
// 全部成功后清理企微已不返回的部门与成员
func (s *ContactSyncer) Sync(ctx context.Context, client *wecom.Client, storeId uint32) error {
	depts, err := client.Contact.ListDepartments(ctx, 0)
	if err != nil {
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsgf("拉取部门列表失败 store_id=%d", storeId))
	}
	if len(depts) == 0 {
		// 拉取源为空时直接返回，不做写入与清理，避免误删本地已有通讯录
		return nil
	}

	// 部门物化路径（父路径+自身ID+冒号，根部门父路径为 "0:"）
	pathMap := buildDeptPathMap(depts)

	// 当前存在的部门ID集合，用于过滤成员数组遗留的已删部门关联
	deptIdSet := make(map[uint32]struct{}, len(depts))
	for _, d := range depts {
		deptIdSet[uint32(d.ID)] = struct{}{}
	}

	members := make(map[string]wecom.User)
	rels := make([]deptMemberRel, 0)
	for _, dept := range depts {
		users, err := client.Contact.ListAllUsersByDepartment(ctx, dept.ID, false)
		if err != nil {
			return apperror.Wrap(errcode.Internal, err, apperror.WithMsgf("拉取部门成员失败 store_id=%d department_id=%d", storeId, dept.ID))
		}
		for _, u := range users {
			if u.UserID == "" {
				continue // 防御：无成员ID的记录不落库
			}
			if _, ok := members[u.UserID]; !ok {
				members[u.UserID] = u
			}
			for i, d := range u.Department {
				did := uint32(d)
				if _, ok := deptIdSet[did]; !ok {
					continue // 防御：成员数组遗留已删部门时跳过该关联
				}
				rels = append(rels, deptMemberRel{
					UserId:       u.UserID,
					DepartmentId: did,
					IsLeader:     indexInt8(u.IsLeaderInDept, i),
					Sort:         indexUint32(u.Order, i),
				})
			}
		}
	}

	if err := s.upsertAll(ctx, storeId, depts, pathMap, members, rels); err != nil {
		return err
	}
	return s.cleanupRemoved(ctx, storeId, depts, members)
}

// buildDeptPathMap 迭代计算各部门物化路径（依赖父路径，多轮直到无进展，防御环）
func buildDeptPathMap(depts []wecom.Department) map[int]string {
	pathMap := make(map[int]string, len(depts))
	remaining := make([]wecom.Department, 0, len(depts))
	for _, d := range depts {
		if d.ParentID == 0 {
			pathMap[d.ID] = "0:" + strconv.Itoa(d.ID) + ":"
		} else {
			remaining = append(remaining, d)
		}
	}
	for len(remaining) > 0 {
		progress := false
		next := remaining[:0]
		for _, d := range remaining {
			if p, ok := pathMap[d.ParentID]; ok {
				pathMap[d.ID] = p + strconv.Itoa(d.ID) + ":"
				progress = true
			} else {
				next = append(next, d)
			}
		}
		remaining = next
		if !progress {
			break // 防御：父级缺失（异常数据）时停止，剩余部门 path 为空
		}
	}
	return pathMap
}

// indexInt8 取数组第 i 个元素，越界返回 0
func indexInt8(list []int, i int) int8 {
	if i >= 0 && i < len(list) {
		return int8(list[i])
	}
	return 0
}

// indexUint32 取数组第 i 个元素，越界返回 0
func indexUint32(list []int, i int) uint32 {
	if i >= 0 && i < len(list) {
		return uint32(list[i])
	}
	return 0
}

// upsertAll 批量差异写：一次读现有数据建映射，分派给部门/成员/关联三个子方法
func (s *ContactSyncer) upsertAll(ctx context.Context, storeId uint32, depts []wecom.Department, pathMap map[int]string,
	members map[string]wecom.User, rels []deptMemberRel) error {

	existingDepts, err := s.deptRepo.FindAll(ctx, &DepartmentFilter{StoreId: storeId}, nil, nil)
	if err != nil {
		return err
	}
	existingMembers, err := s.memberRepo.FindAll(ctx, &MemberFilter{StoreId: storeId}, nil, nil)
	if err != nil {
		return err
	}
	existingRels, err := s.memberDeptRepo.FindAll(ctx, &MemberDepartmentFilter{StoreId: storeId}, nil, nil)
	if err != nil {
		return err
	}

	deptMap := make(map[uint32]*WecomDepartment, len(existingDepts))
	for _, d := range existingDepts {
		deptMap[d.DepartmentId] = d
	}
	memberMap := make(map[string]*WecomMember, len(existingMembers))
	for _, m := range existingMembers {
		memberMap[m.UserId] = m
	}
	relMap := make(map[string]*WecomMemberDepartment, len(existingRels))
	for _, r := range existingRels {
		relMap[relKey(r.UserId, r.DepartmentId)] = r
	}

	if err := s.upsertDepts(ctx, storeId, depts, pathMap, deptMap); err != nil {
		return err
	}
	if err := s.upsertMembers(ctx, storeId, members, memberMap); err != nil {
		return err
	}
	return s.upsertRels(ctx, storeId, rels, relMap)
}

// upsertDepts 部门差异：新增批量插入、有变化按主键更新、无变化跳过
func (s *ContactSyncer) upsertDepts(ctx context.Context, storeId uint32, depts []wecom.Department, pathMap map[int]string, deptMap map[uint32]*WecomDepartment) error {
	toCreate := make([]*WecomDepartment, 0, len(depts))
	toUpdate := make([]*WecomDepartment, 0)
	for _, d := range depts {
		path := pathMap[d.ID]
		model := &WecomDepartment{
			StoreId: storeId, DepartmentId: uint32(d.ID), ParentId: uint32(d.ParentID),
			Path: path, Name: d.Name, Sort: uint32(d.Order),
		}
		if old, ok := deptMap[uint32(d.ID)]; ok {
			model.ID = old.ID
			if old.Path != path || old.ParentId != uint32(d.ParentID) || old.Name != d.Name || old.Sort != uint32(d.Order) {
				toUpdate = append(toUpdate, model)
			}
		} else {
			toCreate = append(toCreate, model)
		}
	}
	if len(toCreate) > 0 {
		if err := s.deptRepo.CreateBatch(ctx, toCreate); err != nil {
			return err
		}
	}
	for _, d := range toUpdate {
		if err := s.deptRepo.Updates(ctx, d, map[string]any{
			"parent_id": d.ParentId, "path": d.Path, "name": d.Name, "sort": d.Sort,
		}); err != nil {
			return err
		}
	}
	return nil
}

// upsertMembers 成员差异：新增批量插入、有变化按主键更新、无变化跳过
func (s *ContactSyncer) upsertMembers(ctx context.Context, storeId uint32, members map[string]wecom.User, memberMap map[string]*WecomMember) error {
	toCreate := make([]*WecomMember, 0, len(members))
	toUpdate := make([]*WecomMember, 0)
	for userId, u := range members {
		model := &WecomMember{
			StoreId: storeId, UserId: userId, Name: u.Name, Position: u.Position,
			Mobile: u.Mobile, Gender: u.Gender, Email: u.Email, Avatar: u.Avatar,
			Status: int8(u.Status),
		}
		if old, ok := memberMap[userId]; ok {
			model.ID = old.ID
			if old.Name != u.Name || old.Position != u.Position || old.Mobile != u.Mobile ||
				old.Gender != u.Gender || old.Email != u.Email || old.Avatar != u.Avatar || old.Status != int8(u.Status) {
				toUpdate = append(toUpdate, model)
			}
		} else {
			toCreate = append(toCreate, model)
		}
	}
	if len(toCreate) > 0 {
		if err := s.memberRepo.CreateBatch(ctx, toCreate); err != nil {
			return err
		}
	}
	for _, m := range toUpdate {
		if err := s.memberRepo.Updates(ctx, m, map[string]any{
			"name": m.Name, "position": m.Position, "mobile": m.Mobile,
			"gender": m.Gender, "email": m.Email, "avatar": m.Avatar, "status": m.Status,
		}); err != nil {
			return err
		}
	}
	return nil
}

// upsertRels 成员部门关联差异：有变化按主键更新、新增批量插入、移除物理删除（关联表无软删字段）
func (s *ContactSyncer) upsertRels(ctx context.Context, storeId uint32, rels []deptMemberRel, relMap map[string]*WecomMemberDepartment) error {
	toCreate := make([]*WecomMemberDepartment, 0)
	toDelete := make([]*WecomMemberDepartment, 0)
	incomingKeys := make(map[string]struct{}, len(rels))
	for _, r := range rels {
		key := relKey(r.UserId, r.DepartmentId)
		incomingKeys[key] = struct{}{}
		if old, ok := relMap[key]; ok {
			if old.IsLeader != r.IsLeader || old.Sort != r.Sort {
				if err := s.memberDeptRepo.Updates(ctx, old, map[string]any{
					"is_leader": r.IsLeader, "sort": r.Sort,
				}); err != nil {
					return err
				}
			}
		} else {
			toCreate = append(toCreate, &WecomMemberDepartment{
				StoreId: storeId, UserId: r.UserId, DepartmentId: r.DepartmentId,
				IsLeader: r.IsLeader, Sort: r.Sort,
			})
		}
	}
	for key, old := range relMap {
		if _, ok := incomingKeys[key]; !ok {
			toDelete = append(toDelete, old)
		}
	}
	if len(toCreate) > 0 {
		if err := s.memberDeptRepo.CreateBatch(ctx, toCreate); err != nil {
			return err
		}
	}
	for _, r := range toDelete {
		if err := s.memberDeptRepo.Delete(ctx, r.ID); err != nil {
			return err
		}
	}
	return nil
}

// relKey 成员部门关联的复合键（user_id 不包含 NUL，安全分隔）
func relKey(userId string, deptId uint32) string {
	return userId + "\x00" + strconv.FormatUint(uint64(deptId), 10)
}

// cleanupRemoved 清理已移除数据：企微不再返回的部门与成员软删（成员部门关联无软删字段，
// 部门软删后经 join 查询自然不可见，不单独清理）。仅在整轮拉取成功后执行
func (s *ContactSyncer) cleanupRemoved(ctx context.Context, storeId uint32, depts []wecom.Department, members map[string]wecom.User) error {
	deptSeen := make(map[uint32]struct{}, len(depts))
	for _, d := range depts {
		deptSeen[uint32(d.ID)] = struct{}{}
	}
	existingDepts, err := s.deptRepo.FindAll(ctx, &DepartmentFilter{StoreId: storeId}, nil, nil)
	if err != nil {
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsgf("查询本地部门失败 store_id=%d", storeId))
	}
	for _, d := range existingDepts {
		if _, ok := deptSeen[d.DepartmentId]; !ok {
			if err := s.deptRepo.Delete(ctx, d.ID); err != nil {
				return apperror.Wrap(errcode.Internal, err, apperror.WithMsgf("清理已移除部门失败 department_id=%d", d.DepartmentId))
			}
		}
	}

	existingMembers, err := s.memberRepo.FindAll(ctx, &MemberFilter{StoreId: storeId}, nil, nil)
	if err != nil {
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsgf("查询本地成员失败 store_id=%d", storeId))
	}
	for _, m := range existingMembers {
		if _, ok := members[m.UserId]; !ok {
			if err := s.memberRepo.Delete(ctx, m.ID); err != nil {
				return apperror.Wrap(errcode.Internal, err, apperror.WithMsgf("清理已移除成员失败 user_id=%s", m.UserId))
			}
		}
	}
	return nil
}
