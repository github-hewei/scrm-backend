package config

import (
	"github.com/241x/zero-kit/baserepo"
	"gorm.io/gorm"
)

// AppType 应用类型
type AppType int8

// 应用类型定义
const (
	AppTypeSelfBuilt   AppType = 10 // 自建应用
	AppTypeContact     AppType = 20 // 客户联系
	AppTypeAddressBook AppType = 30 // 通讯录同步
)

// Name 应用类型中文名称
func (t AppType) Name() string {
	switch t {
	case AppTypeSelfBuilt:
		return "自建应用"
	case AppTypeContact:
		return "客户联系"
	case AppTypeAddressBook:
		return "通讯录同步"
	default:
		return ""
	}
}

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
	Status  int8
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
	if f.Status != 0 {
		db = db.Where("status = ?", f.Status)
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
