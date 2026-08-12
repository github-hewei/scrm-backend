package contact

import (
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
	StoreId      uint32
	DepartmentId uint32
	ParentId     uint32
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

// MemberFilter 成员过滤条件
type MemberFilter struct {
	StoreId uint32
	UserId  string
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

// MemberDepartmentFilter 成员部门关联过滤条件
type MemberDepartmentFilter struct {
	StoreId      uint32
	UserId       string
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
	if f.DepartmentId != 0 {
		db = db.Where("department_id = ?", f.DepartmentId)
	}
	return db
}
