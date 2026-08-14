package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"zero-backend/internal/modules/async"
	wecomsync "zero-backend/internal/modules/wecom/sync"

	"github.com/241x/zero-kit/job"
	"github.com/241x/zero-kit/logger"
)

// WecomSyncResult 同步作业结果（单企业粒度）
type WecomSyncResult struct {
	StoreId uint32 `json:"store_id"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// WecomSyncJobHandler 企业微信数据同步作业处理器：与 CLI runner 共用同步内核，
// 并通过 async.Recorder 同步业务任务记录（租户展示）
type WecomSyncJobHandler struct {
	svc     *wecomsync.Service
	taskSvc *async.TaskService
	log     logger.Logger
}

// NewWecomSyncJobHandler 创建同步作业处理器
func NewWecomSyncJobHandler(svc *wecomsync.Service, taskSvc *async.TaskService, log logger.Logger) *WecomSyncJobHandler {
	return &WecomSyncJobHandler{svc: svc, taskSvc: taskSvc, log: log}
}

// Execute 执行同步作业：解析 payload → 批量同步 → 同步任务状态/进度/结果。
// 失败返回 error 后由执行器按 MaxAttempts 重试；同步本身幂等，重复执行安全
func (h *WecomSyncJobHandler) Execute(ctx context.Context, j *job.Job) (retErr error) {
	if j.Type != wecomsync.JobTypeWecomSync {
		return fmt.Errorf("不支持的作业类型: %s", j.Type)
	}

	var req wecomsync.WecomSyncJobPayload
	if err := json.Unmarshal(j.Payload, &req); err != nil {
		return fmt.Errorf("解析同步作业参数失败: %w", err)
	}
	scope, err := wecomsync.ParseScope(req.Scope)
	if err != nil {
		return err
	}

	// 业务任务记录同步（记录不存在时静默忽略）
	rec := async.NewRecorder(h.taskSvc, j.ID)
	if err := rec.Start(ctx); err != nil {
		h.log.Warn("同步任务启动记录失败", "job_id", j.ID, "error", err)
	}

	start := time.Now()
	storeIds, err := h.svc.LoadStoreIds(ctx, req.StoreId)
	if err != nil {
		return err
	}
	if len(storeIds) == 0 {
		return fmt.Errorf("没有需要同步的企业")
	}

	results := make([]WecomSyncResult, 0, len(storeIds))
	onProgress := func(storeId uint32, completed, total int, syncErr error) {
		result := WecomSyncResult{StoreId: storeId, Success: syncErr == nil}
		if syncErr != nil {
			result.Error = syncErr.Error()
		}
		results = append(results, result)
		progress := completed * 100 / total
		job.ReportProgress(ctx, progress)
		if err := rec.Progress(ctx, progress); err != nil {
			h.log.Warn("同步任务进度记录失败", "job_id", j.ID, "error", err)
		}
	}

	retErr = h.svc.SyncStores(ctx, storeIds, scope, onProgress)

	// 作业被取消（ctx 已关闭）时任务标记为 cancelled，与失败区分开
	if errors.Is(retErr, context.Canceled) {
		if err := rec.Cancel(ctx); err != nil {
			h.log.Warn("同步任务取消记录失败", "job_id", j.ID, "error", err)
		}
		return retErr
	}

	resultJSON, _ := json.Marshal(map[string]any{
		"scope":   req.Scope,
		"total":   len(storeIds),
		"results": results,
		"cost_ms": time.Since(start).Milliseconds(),
	})
	j.Result = resultJSON
	if err := rec.Finish(ctx, retErr, resultJSON); err != nil {
		h.log.Warn("同步任务结果记录失败", "job_id", j.ID, "error", err)
	}
	return retErr
}
