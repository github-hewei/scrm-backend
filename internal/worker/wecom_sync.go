package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

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

// WecomSyncJobHandler 企业微信数据同步作业处理器：与 CLI runner 共用同步内核
type WecomSyncJobHandler struct {
	svc *wecomsync.Service
	log logger.Logger
}

// NewWecomSyncJobHandler 创建同步作业处理器
func NewWecomSyncJobHandler(svc *wecomsync.Service, log logger.Logger) *WecomSyncJobHandler {
	return &WecomSyncJobHandler{svc: svc, log: log}
}

// Execute 执行同步作业：解析 payload → 批量同步 → 上报进度 → 写入结果。
// 失败返回 error 后由执行器按 MaxAttempts 重试；同步本身幂等，重复执行安全
func (h *WecomSyncJobHandler) Execute(ctx context.Context, j *job.Job) error {
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
		job.ReportProgress(ctx, completed*100/total)
	}

	err = h.svc.SyncStores(ctx, storeIds, scope, onProgress)

	resultJSON, _ := json.Marshal(map[string]any{
		"scope":   req.Scope,
		"total":   len(storeIds),
		"results": results,
		"cost_ms": time.Since(start).Milliseconds(),
	})
	j.Result = resultJSON
	return err
}
