package async

import "gorm.io/plugin/soft_delete"

// TaskStatus 异步任务状态
type TaskStatus string

// 任务状态定义（与 job 包状态语义对齐，业务侧独立）
const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusSuccess   TaskStatus = "success"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCancelled TaskStatus = "cancelled"
)

// AsyncTask 通用异步任务记录：接口调用时不方便同步执行的任务，
// 供租户前端展示状态/进度/结果摘要；调度执行由系统级 jobs 表承载，job_id 单向关联
type AsyncTask struct {
	ID        uint64 `json:"id" gorm:"primaryKey"`
	StoreId   uint32 `json:"store_id" gorm:"not null;default:0;comment:企业ID(租户)，0=系统级任务;index:idx_store_status,priority:1;index:idx_store_type,priority:1"`
	JobId     string `json:"job_id" gorm:"size:64;not null;default:'';comment:关联系统调度表jobs.id(1:1);uniqueIndex"`
	TaskType  string `json:"task_type" gorm:"size:64;not null;default:'';comment:任务类型(wecom.sync/mail.send/export等);index:idx_store_type,priority:2"`
	Title     string `json:"title" gorm:"size:255;not null;default:'';comment:租户友好标题"`
	Status    TaskStatus `json:"status" gorm:"size:16;not null;default:'pending';comment:任务状态(pending/running/success/failed/cancelled);index:idx_store_status,priority:2"`
	Progress  int    `json:"progress" gorm:"not null;default:0;comment:执行进度(0-100)"`
	Error     string `json:"error" gorm:"size:2000;not null;default:'';comment:最近一次错误信息"`
	Result    []byte `json:"result" gorm:"type:json;comment:结果摘要(各类型自定义JSON)"`
	CreatedAt uint32 `json:"created_at" gorm:"not null;comment:创建时间;autoCreateTime"`
	UpdatedAt uint32 `json:"updated_at" gorm:"not null;comment:更新时间;autoUpdateTime"`

	DeletedAt soft_delete.DeletedAt `json:"-" gorm:"not null;default:0;comment:删除时间"`
}
