package setting

import (
	"context"

	"github.com/241x/zero-kit/baserepo"
	"gorm.io/gorm"
)

// Filter 过滤条件
type Filter struct {
	Id         uint32
	StoreId    uint32
	SettingKey string
}

// Apply 应用过滤条件
func (f *Filter) Apply(db *gorm.DB) *gorm.DB {
	if f == nil {
		return db
	}
	if f.Id != 0 {
		db = db.Where("id = ?", f.Id)
	}
	if f.StoreId != 0 {
		db = db.Where("store_id = ?", f.StoreId)
	}
	if f.SettingKey != "" {
		db = db.Where("setting_key = ?", f.SettingKey)
	}
	return db
}

// Repository 数据仓库
type Repository struct {
	*baserepo.BaseRepository[Setting]
}

// NewRepository 创建数据仓库
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{BaseRepository: baserepo.NewBaseRepository[Setting](db)}
}

// FindByKeyAndStore 按key和store_id精确查询，storeId=0时也会过滤
func (r *Repository) FindByKeyAndStore(ctx context.Context, key string, storeId uint32) (*Setting, error) {
	var setting Setting
	err := r.Db.WithContext(ctx).Where("setting_key = ? AND store_id = ?", key, storeId).First(&setting).Error
	if err != nil {
		return nil, err
	}
	return &setting, nil
}

// DefaultFilter 默认过滤条件
type DefaultFilter struct {
	SettingKey string
}

// Apply 应用过滤条件
func (f *DefaultFilter) Apply(db *gorm.DB) *gorm.DB {
	if f == nil {
		return db
	}
	if f.SettingKey != "" {
		db = db.Where("setting_key = ?", f.SettingKey)
	}
	return db
}

// DefaultRepository 默认数据仓库
type DefaultRepository struct {
	*baserepo.BaseRepository[SettingDefault]
}

// NewDefaultRepository 创建默认数据仓库
func NewDefaultRepository(db *gorm.DB) *DefaultRepository {
	return &DefaultRepository{BaseRepository: baserepo.NewBaseRepository[SettingDefault](db)}
}
