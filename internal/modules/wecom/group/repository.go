package group

import (
	"github.com/241x/zero-kit/baserepo"
	"gorm.io/gorm"
)

// WecomGroupRepository 企业微信客户群仓库
type WecomGroupRepository struct {
	*baserepo.BaseRepository[WecomGroup]
}

// NewWecomGroupRepository 创建客户群仓库
func NewWecomGroupRepository(db *gorm.DB) *WecomGroupRepository {
	return &WecomGroupRepository{BaseRepository: baserepo.NewBaseRepository[WecomGroup](db)}
}

// GroupFilter 客户群过滤条件
type GroupFilter struct {
	StoreId uint32
	ChatId  string
	Owner   string
	Status  int8
}

// Apply 应用过滤条件
func (f *GroupFilter) Apply(db *gorm.DB) *gorm.DB {
	if f == nil {
		return db
	}
	if f.StoreId != 0 {
		db = db.Where("store_id = ?", f.StoreId)
	}
	if f.ChatId != "" {
		db = db.Where("chat_id = ?", f.ChatId)
	}
	if f.Owner != "" {
		db = db.Where("owner = ?", f.Owner)
	}
	if f.Status != 0 {
		db = db.Where("status = ?", f.Status)
	}
	return db
}

// WecomGroupMemberRepository 客户群成员仓库
type WecomGroupMemberRepository struct {
	*baserepo.BaseRepository[WecomGroupMember]
}

// NewWecomGroupMemberRepository 创建客户群成员仓库
func NewWecomGroupMemberRepository(db *gorm.DB) *WecomGroupMemberRepository {
	return &WecomGroupMemberRepository{BaseRepository: baserepo.NewBaseRepository[WecomGroupMember](db)}
}

// GroupMemberFilter 客户群成员过滤条件
type GroupMemberFilter struct {
	StoreId uint32
	ChatId  string
	UserId  string
	Type    int8
}

// Apply 应用过滤条件
func (f *GroupMemberFilter) Apply(db *gorm.DB) *gorm.DB {
	if f == nil {
		return db
	}
	if f.StoreId != 0 {
		db = db.Where("store_id = ?", f.StoreId)
	}
	if f.ChatId != "" {
		db = db.Where("chat_id = ?", f.ChatId)
	}
	if f.UserId != "" {
		db = db.Where("user_id = ?", f.UserId)
	}
	if f.Type != 0 {
		db = db.Where("type = ?", f.Type)
	}
	return db
}
