package contact

import (
	"context"

	"github.com/241x/zero-kit/baserepo"
	"gorm.io/gorm"
)

// WecomDepartmentRepository 企业微信部门仓库
type WecomDepartmentRepository struct {
	*baserepo.BaseRepository[WecomDepartment]
}

// NewWecomDepartmentRepository 创建部门仓库
func NewWecomDepartmentRepository(db *gorm.DB) *WecomDepartmentRepository {
	return &WecomDepartmentRepository{BaseRepository: baserepo.NewBaseRepository[WecomDepartment](db)}
}

// DepartmentFilter 部门过滤条件
type DepartmentFilter struct {
	StoreId       uint32
	DepartmentId  uint32
	DepartmentIds []uint32
	ParentId      uint32
}

// Apply 应用过滤条件
func (f *DepartmentFilter) Apply(db *gorm.DB) *gorm.DB {
	if f == nil {
		return db
	}
	if f.StoreId != 0 {
		db = db.Where("store_id = ?", f.StoreId)
	}
	if f.DepartmentId != 0 {
		db = db.Where("department_id = ?", f.DepartmentId)
	}
	if len(f.DepartmentIds) > 0 {
		db = db.Where("department_id IN ?", f.DepartmentIds)
	}
	if f.ParentId != 0 {
		db = db.Where("parent_id = ?", f.ParentId)
	}
	return db
}

// WecomMemberRepository 企业微信成员仓库
type WecomMemberRepository struct {
	*baserepo.BaseRepository[WecomMember]
}

// NewWecomMemberRepository 创建成员仓库
func NewWecomMemberRepository(db *gorm.DB) *WecomMemberRepository {
	return &WecomMemberRepository{BaseRepository: baserepo.NewBaseRepository[WecomMember](db)}
}

// FindMemberNames 按成员ID列表批量查询姓名，返回 user_id -> name 映射（跨包供客户模块调用）
func (r *WecomMemberRepository) FindMemberNames(ctx context.Context, storeId uint32, userIds []string) (map[string]string, error) {
	result := make(map[string]string, len(userIds))
	if len(userIds) == 0 {
		return result, nil
	}
	list, err := r.FindAll(ctx, &MemberFilter{StoreId: storeId, UserIds: userIds}, nil, nil)
	if err != nil {
		return nil, err
	}
	for _, member := range list {
		result[member.UserId] = member.Name
	}
	return result, nil
}

// FindPageByDepartmentPath 按部门物化路径分页查询成员（含所有子孙部门），返回列表与总数。
// 使用子查询+semi-join避免大IN列表，departmentPath 为部门自身的 path（如 0:1:）
func (r *WecomMemberRepository) FindPageByDepartmentPath(ctx context.Context, storeId uint32, departmentPath string, status int8, page, limit int) ([]*WecomMember, int64, error) {
	sub := r.Db.WithContext(ctx).Table("gaz_wecom_member_department md").
		Select("md.user_id").
		Joins("JOIN gaz_wecom_department d ON md.department_id = d.department_id").
		Where("md.store_id = ? AND d.path LIKE ?", storeId, departmentPath+"%")

	countQuery := r.Db.WithContext(ctx).Model(new(WecomMember)).
		Where("store_id = ?", storeId).
		Where("user_id IN (?)", sub)
	if status != 0 {
		countQuery = countQuery.Where("status = ?", status)
	}
	var total int64
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	query := r.Db.WithContext(ctx).Model(new(WecomMember)).
		Where("store_id = ?", storeId).
		Where("user_id IN (?)", sub)
	if status != 0 {
		query = query.Where("status = ?", status)
	}
	list := make([]*WecomMember, 0)
	if err := query.Order("id asc").Limit(limit).Offset((page - 1) * limit).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// MemberFilter 成员过滤条件
type MemberFilter struct {
	StoreId uint32
	UserId  string
	UserIds []string
	Status  int8
}

// Apply 应用过滤条件
func (f *MemberFilter) Apply(db *gorm.DB) *gorm.DB {
	if f == nil {
		return db
	}
	if f.StoreId != 0 {
		db = db.Where("store_id = ?", f.StoreId)
	}
	if f.UserId != "" {
		db = db.Where("user_id = ?", f.UserId)
	}
	if len(f.UserIds) > 0 {
		db = db.Where("user_id IN ?", f.UserIds)
	}
	if f.Status != 0 {
		db = db.Where("status = ?", f.Status)
	}
	return db
}

// WecomMemberDepartmentRepository 成员部门关联仓库
type WecomMemberDepartmentRepository struct {
	*baserepo.BaseRepository[WecomMemberDepartment]
}

// NewWecomMemberDepartmentRepository 创建成员部门关联仓库
func NewWecomMemberDepartmentRepository(db *gorm.DB) *WecomMemberDepartmentRepository {
	return &WecomMemberDepartmentRepository{BaseRepository: baserepo.NewBaseRepository[WecomMemberDepartment](db)}
}

// CountByDepartment 按部门统计成员数，返回 department_id -> 成员数 映射
func (r *WecomMemberDepartmentRepository) CountByDepartment(ctx context.Context, storeId uint32) (map[uint32]uint32, error) {
	rows := make([]struct {
		DepartmentId uint32
		Cnt          uint32
	}, 0)
	if err := r.Db.WithContext(ctx).Model(new(WecomMemberDepartment)).
		Select("department_id, COUNT(*) AS cnt").
		Where("store_id = ?", storeId).
		Group("department_id").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	result := make(map[uint32]uint32, len(rows))
	for _, row := range rows {
		result[row.DepartmentId] = row.Cnt
	}
	return result, nil
}

// FindListByUserIds 按成员ID列表查询部门关联（列表页补部门信息用，避免全量拉取）
func (r *WecomMemberDepartmentRepository) FindListByUserIds(ctx context.Context, storeId uint32, userIds []string) ([]*WecomMemberDepartment, error) {
	if len(userIds) == 0 {
		return []*WecomMemberDepartment{}, nil
	}
	return r.FindAll(ctx, &MemberDepartmentFilter{StoreId: storeId, UserIds: userIds}, nil, nil)
}

// MemberDepartmentFilter 成员部门关联过滤条件
type MemberDepartmentFilter struct {
	StoreId      uint32
	UserId       string
	UserIds      []string
	DepartmentId uint32
}

// Apply 应用过滤条件
func (f *MemberDepartmentFilter) Apply(db *gorm.DB) *gorm.DB {
	if f == nil {
		return db
	}
	if f.StoreId != 0 {
		db = db.Where("store_id = ?", f.StoreId)
	}
	if f.UserId != "" {
		db = db.Where("user_id = ?", f.UserId)
	}
	if len(f.UserIds) > 0 {
		db = db.Where("user_id IN ?", f.UserIds)
	}
	if f.DepartmentId != 0 {
		db = db.Where("department_id = ?", f.DepartmentId)
	}
	return db
}
