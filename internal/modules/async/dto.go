package async

import (
	"encoding/json"
)

// SubmitTaskRequest 提交异步任务请求（store_id 取自登录上下文，不接收请求体传入）
type SubmitTaskRequest struct {
	TaskType string          `json:"task_type" validate:"required,max=64"`
	Payload  json.RawMessage `json:"payload"` // 任务入参，由各任务类型的注册规则自行解析
}

// SubmitTaskResponse 提交异步任务响应
type SubmitTaskResponse struct {
	TaskId uint64 `json:"task_id"`
}

// TaskListRequest 任务列表请求
type TaskListRequest struct {
	Page     int    `json:"page" validate:"required,min=1"`
	Limit    int    `json:"limit" validate:"required,min=1,max=100"`
	TaskType string `json:"task_type"` // 任务类型过滤，空=全部
}

// TaskListResponse 任务列表响应
type TaskListResponse struct {
	List  []*TaskInfo `json:"list"`
	Total int64       `json:"total"`
}

// TaskDetailRequest 任务详情请求
type TaskDetailRequest struct {
	Id uint64 `json:"id" validate:"required,min=1"`
}

// TaskInfo 任务展示信息（租户友好视图，不暴露 jobs 内部字段）
type TaskInfo struct {
	Id        uint64          `json:"id"`
	TaskType  string          `json:"task_type"`
	Title     string          `json:"title"`
	Status    TaskStatus      `json:"status"`
	Progress  int             `json:"progress"`
	Error     string          `json:"error"`
	Result    json.RawMessage `json:"result"`
	CreatedAt uint32          `json:"created_at"`
	UpdatedAt uint32          `json:"updated_at"`
}

// fromTask 任务模型转展示 DTO
func fromTask(t *AsyncTask) *TaskInfo {
	return &TaskInfo{
		Id:        t.ID,
		TaskType:  t.TaskType,
		Title:     t.Title,
		Status:    t.Status,
		Progress:  t.Progress,
		Error:     t.Error,
		Result:    t.Result,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}

// fromTasks 任务列表转展示 DTO 列表
func fromTasks(list []*AsyncTask) []*TaskInfo {
	out := make([]*TaskInfo, 0, len(list))
	for _, t := range list {
		out = append(out, fromTask(t))
	}
	return out
}
