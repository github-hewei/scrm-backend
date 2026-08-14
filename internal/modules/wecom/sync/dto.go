package sync

import (
	"encoding/json"
	"strconv"

	"github.com/241x/zero-kit/job"
)

// SubmitSyncRequest 提交同步作业请求（store_id 取自登录上下文，不接收请求体传入）
type SubmitSyncRequest struct {
	Scope string `json:"scope" validate:"required,oneof=all dept contact group"`
}

// SubmitSyncResponse 提交同步作业响应
type SubmitSyncResponse struct {
	JobId string `json:"job_id"`
}

// JobListRequest 同步作业列表请求
type JobListRequest struct {
	Page  int `json:"page" validate:"required,min=1"`
	Limit int `json:"limit" validate:"required,min=1,max=100"`
}

// JobListResponse 同步作业列表响应
type JobListResponse struct {
	List  []*JobDetail `json:"list"`
	Total int64        `json:"total"`
}

// JobDetailRequest 同步作业详情请求
type JobDetailRequest struct {
	JobId string `json:"job_id" validate:"required,max=64"`
}

// JobDetail 同步作业详情
type JobDetail struct {
	ID          string          `json:"id"`
	Type        string          `json:"type"`
	Status      job.Status      `json:"status"`
	Progress    int             `json:"progress"`
	Error       string          `json:"error"`
	Payload     json.RawMessage `json:"payload"`
	Result      json.RawMessage `json:"result"`
	StoreId     uint32          `json:"store_id"`
	Attempts    int             `json:"attempts"`
	MaxAttempts int             `json:"max_attempts"`
	CreatedAt   int64           `json:"created_at"`
	StartedAt   int64           `json:"started_at"`
	CompletedAt int64           `json:"completed_at"`
}

// fromJob 作业模型转详情 DTO
func fromJob(j *job.Job) *JobDetail {
	storeId, _ := strconv.ParseUint(j.Metadata[metadataStoreId], 10, 32)
	return &JobDetail{
		ID:          j.ID,
		Type:        j.Type,
		Status:      j.Status,
		Progress:    j.Progress,
		Error:       j.Error,
		Payload:     j.Payload,
		Result:      j.Result,
		StoreId:     uint32(storeId),
		Attempts:    j.Attempts,
		MaxAttempts: j.MaxAttempts,
		CreatedAt:   j.CreatedAt,
		StartedAt:   j.StartedAt,
		CompletedAt: j.CompletedAt,
	}
}

// fromJobs 作业列表转详情 DTO 列表
func fromJobs(list []*job.Job) []*JobDetail {
	out := make([]*JobDetail, 0, len(list))
	for _, j := range list {
		out = append(out, fromJob(j))
	}
	return out
}
