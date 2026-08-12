package customer

import (
	"github.com/241x/zero-kit/baserepo"
	"gorm.io/gorm"
)

// WecomCustomerRepository 企业微信客户仓库
type WecomCustomerRepository struct {
	*baserepo.BaseRepository[WecomCustomer]
}

// NewWecomCustomerRepository 创建客户仓库
func NewWecomCustomerRepository(db *gorm.DB) *WecomCustomerRepository {
	return &WecomCustomerRepository{BaseRepository: baserepo.NewBaseRepository[WecomCustomer](db)}
}

// CustomerFilter 客户过滤条件
type CustomerFilter struct {
	StoreId        uint32
	ExternalUserid string
	Unionid        string
	Type           int8
}

// Apply 应用过滤条件
func (f *CustomerFilter) Apply(db *gorm.DB) *gorm.DB {
	if f == nil {
		return db
	}
	if f.StoreId != 0 {
		db = db.Where("store_id = ?", f.StoreId)
	}
	if f.ExternalUserid != "" {
		db = db.Where("external_userid = ?", f.ExternalUserid)
	}
	if f.Unionid != "" {
		db = db.Where("unionid = ?", f.Unionid)
	}
	if f.Type != 0 {
		db = db.Where("type = ?", f.Type)
	}
	return db
}

// WecomCustomerFollowRepository 客户跟进仓库
type WecomCustomerFollowRepository struct {
	*baserepo.BaseRepository[WecomCustomerFollow]
}

// NewWecomCustomerFollowRepository 创建客户跟进仓库
func NewWecomCustomerFollowRepository(db *gorm.DB) *WecomCustomerFollowRepository {
	return &WecomCustomerFollowRepository{BaseRepository: baserepo.NewBaseRepository[WecomCustomerFollow](db)}
}

// CustomerFollowFilter 客户跟进过滤条件
type CustomerFollowFilter struct {
	StoreId        uint32
	ExternalUserid string
	UserId         string
	State          string
}

// Apply 应用过滤条件
func (f *CustomerFollowFilter) Apply(db *gorm.DB) *gorm.DB {
	if f == nil {
		return db
	}
	if f.StoreId != 0 {
		db = db.Where("store_id = ?", f.StoreId)
	}
	if f.ExternalUserid != "" {
		db = db.Where("external_userid = ?", f.ExternalUserid)
	}
	if f.UserId != "" {
		db = db.Where("user_id = ?", f.UserId)
	}
	if f.State != "" {
		db = db.Where("state = ?", f.State)
	}
	return db
}
