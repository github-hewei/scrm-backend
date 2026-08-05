package upload

import (
	"github.com/241x/zero-kit/baserepo"
	"github.com/241x/zero-kit/helper"
	"gorm.io/gorm"
)

// GroupFilter 文件分组筛选字段
type GroupFilter struct {
	Id      uint32
	StoreId uint32
	Name    string
}

// Apply 应用过滤条件
func (f *GroupFilter) Apply(db *gorm.DB) *gorm.DB {
	if f == nil {
		return db
	}
	if f.Id != 0 {
		db = db.Where("id = ?", f.Id)
	}
	db = db.Where("store_id = ?", f.StoreId)
	if f.Name != "" {
		db = db.Where("name = ?", f.Name)
	}
	return db
}

// GroupRepository 文件分组仓库
type GroupRepository struct {
	*baserepo.BaseRepository[UploadGroup]
}

// NewGroupRepository 创建文件分组仓库
func NewGroupRepository(db *gorm.DB) *GroupRepository {
	return &GroupRepository{BaseRepository: baserepo.NewBaseRepository[UploadGroup](db)}
}

// FileFilter 文件查询过滤字段
type FileFilter struct {
	Id       uint32
	StoreId  uint32
	GroupId  string
	FileType int8
	FileName string
}

// Apply 应用过滤条件
func (f *FileFilter) Apply(db *gorm.DB) *gorm.DB {
	if f == nil {
		return db
	}
	if f.Id != 0 {
		db = db.Where("id = ?", f.Id)
	}
	db = db.Where("store_id = ?", f.StoreId)
	if f.GroupId != "" && f.GroupId != "all" {
		db = db.Where("group_id = ?", f.GroupId)
	}
	if f.FileType > 0 {
		db = db.Where("file_type = ?", f.FileType)
	}
	if f.FileName != "" {
		db = db.Where("file_name like ?", helper.SafeLikeString(f.FileName)+"%")
	}
	return db
}

// FileRepository 文件仓库
type FileRepository struct {
	*baserepo.BaseRepository[UploadFile]
}

// NewFileRepository 创建文件仓库
func NewFileRepository(db *gorm.DB) *FileRepository {
	return &FileRepository{BaseRepository: baserepo.NewBaseRepository[UploadFile](db)}
}
