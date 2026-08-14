package async

import (
	"github.com/241x/zero-kit/baserepo"
	"gorm.io/gorm"
)

// AsyncTaskRepository 通用异步任务仓库
type AsyncTaskRepository struct {
	*baserepo.BaseRepository[AsyncTask]
}

// NewAsyncTaskRepository 创建异步任务仓库
func NewAsyncTaskRepository(db *gorm.DB) *AsyncTaskRepository {
	return &AsyncTaskRepository{BaseRepository: baserepo.NewBaseRepository[AsyncTask](db)}
}

// AsyncTaskFilter 异步任务过滤条件
type AsyncTaskFilter struct {
	Id       uint64
	StoreId  uint32
	JobId    string
	TaskType string
	Status   string
	Statuses []string
}

// Apply 应用过滤条件
func (f *AsyncTaskFilter) Apply(db *gorm.DB) *gorm.DB {
	if f == nil {
		return db
	}
	if f.Id != 0 {
		db = db.Where("id = ?", f.Id)
	}
	if f.StoreId != 0 {
		db = db.Where("store_id = ?", f.StoreId)
	}
	if f.JobId != "" {
		db = db.Where("job_id = ?", f.JobId)
	}
	if f.TaskType != "" {
		db = db.Where("task_type = ?", f.TaskType)
	}
	if f.Status != "" {
		db = db.Where("status = ?", f.Status)
	}
	if len(f.Statuses) > 0 {
		db = db.Where("status IN ?", f.Statuses)
	}
	return db
}
