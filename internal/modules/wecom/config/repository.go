package config

import (
	"github.com/241x/zero-kit/baserepo"
	"gorm.io/gorm"
)

// WecomConfigRepository 企业微信配置仓库
type WecomConfigRepository struct {
	*baserepo.BaseRepository[WecomConfig]
}

// NewWecomConfigRepository 创建配置仓库
func NewWecomConfigRepository(db *gorm.DB) *WecomConfigRepository {
	return &WecomConfigRepository{BaseRepository: baserepo.NewBaseRepository[WecomConfig](db)}
}

// WecomConfigFilter 配置过滤条件
type WecomConfigFilter struct {
	Id      uint32
	StoreId uint32
	CorpId  string
}

// Apply 应用过滤条件
func (f *WecomConfigFilter) Apply(db *gorm.DB) *gorm.DB {
	if f == nil {
		return db
	}
	if f.Id != 0 {
		db = db.Where("id = ?", f.Id)
	}
	if f.StoreId != 0 {
		db = db.Where("store_id = ?", f.StoreId)
	}
	if f.CorpId != "" {
		db = db.Where("corp_id = ?", f.CorpId)
	}
	return db
}

// WecomAppRepository 企业微信应用仓库
type WecomAppRepository struct {
	*baserepo.BaseRepository[WecomApp]
}

// NewWecomAppRepository 创建应用仓库
func NewWecomAppRepository(db *gorm.DB) *WecomAppRepository {
	return &WecomAppRepository{BaseRepository: baserepo.NewBaseRepository[WecomApp](db)}
}

// WecomAppFilter 应用过滤条件
type WecomAppFilter struct {
	Id            uint32
	StoreId       uint32
	AppType       int8
	AgentId       uint32
	CallbackToken string
}

// Apply 应用过滤条件
func (f *WecomAppFilter) Apply(db *gorm.DB) *gorm.DB {
	if f == nil {
		return db
	}
	if f.Id != 0 {
		db = db.Where("id = ?", f.Id)
	}
	if f.StoreId != 0 {
		db = db.Where("store_id = ?", f.StoreId)
	}
	if f.AppType != 0 {
		db = db.Where("app_type = ?", f.AppType)
	}
	if f.AgentId != 0 {
		db = db.Where("agent_id = ?", f.AgentId)
	}
	if f.CallbackToken != "" {
		db = db.Where("callback_token = ?", f.CallbackToken)
	}
	return db
}
