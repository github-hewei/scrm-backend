package sync

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"

	wecomconfig "zero-backend/internal/modules/wecom/config"

	"github.com/241x/zero-kit/apperror"
	"github.com/241x/zero-kit/baserepo"
	"github.com/241x/zero-kit/job"
	"github.com/241x/zero-web/errcode"
)

// JobTypeWecomSync 企业微信数据同步作业类型（worker 处理器按此注册）
const JobTypeWecomSync = "wecom.sync"

// metadataStoreId 作业元数据中企业ID的键（用于租户隔离过滤）
const metadataStoreId = "store_id"

// WecomSyncJobPayload 同步作业入参
type WecomSyncJobPayload struct {
	StoreId uint32 `json:"store_id"` // 企业ID，0=全部已接入企业
	Scope   string `json:"scope"`    // all|dept|contact|group
}

// JobService 同步作业管理服务（web 端触发/查询）
type JobService struct {
	configRepo  *wecomconfig.WecomConfigRepository
	store       *job.SQLStore
	maxAttempts int
}

// NewJobService 创建同步作业管理服务
func NewJobService(configRepo *wecomconfig.WecomConfigRepository, store *job.SQLStore) *JobService {
	return &JobService{configRepo: configRepo, store: store, maxAttempts: 3}
}

// Submit 提交同步作业：校验企业已接入、scope 合法且无执行中的重复作业，返回作业ID。
// 提交成功后由 worker 进程的作业执行器异步执行
func (s *JobService) Submit(ctx context.Context, storeId uint32, scope string) (string, error) {
	if _, err := ParseScope(scope); err != nil {
		return "", apperror.New(errcode.InvalidInput, apperror.WithMsg(err.Error()))
	}
	if storeId == 0 {
		return "", apperror.New(errcode.InvalidInput, apperror.WithMsg("企业标识缺失"))
	}
	if _, err := s.configRepo.FindOne(ctx, &wecomconfig.WecomConfigFilter{StoreId: storeId, Status: 1}); err != nil {
		if errors.Is(err, baserepo.ErrRecordNotFound) {
			return "", apperror.New(errcode.NotFound, apperror.WithMsgf("企业未接入或已停用 store_id=%d", storeId))
		}
		return "", apperror.Wrap(errcode.Internal, err, apperror.WithMsg("校验企业接入配置失败"))
	}

	// 连点去重：该企业已有 pending/running 作业时拒绝重复提交
	existing, err := s.store.List(ctx, job.JobFilter{Queue: job.DefaultQueue, Type: JobTypeWecomSync, Limit: 100})
	if err != nil {
		return "", apperror.Wrap(errcode.Internal, err, apperror.WithMsg("查询进行中作业失败"))
	}
	for _, j := range existing {
		if j.Status != job.StatusPending && j.Status != job.StatusRunning {
			continue
		}
		var p WecomSyncJobPayload
		if json.Unmarshal(j.Payload, &p) == nil && p.StoreId == storeId {
			return "", apperror.New(errcode.Conflict, apperror.WithMsgf("该企业已有同步作业在执行 job_id=%s", j.ID))
		}
	}

	payload, _ := json.Marshal(WecomSyncJobPayload{StoreId: storeId, Scope: scope})
	j := job.NewJob(JobTypeWecomSync, payload).
		WithMaxAttempts(s.maxAttempts).
		WithMetadata(metadataStoreId, strconv.FormatUint(uint64(storeId), 10))
	if err := s.store.Save(ctx, j); err != nil {
		return "", apperror.Wrap(errcode.Internal, err, apperror.WithMsg("提交同步作业失败"))
	}
	return j.ID, nil
}

// Get 查询作业详情（校验归属当前企业，防止跨企业访问）
func (s *JobService) Get(ctx context.Context, storeId uint32, jobID string) (*job.Job, error) {
	j, err := s.store.Get(ctx, jobID)
	if err != nil {
		if errors.Is(err, job.ErrJobNotFound) {
			return nil, apperror.New(errcode.NotFound, apperror.WithMsg("作业不存在"))
		}
		return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("查询作业失败"))
	}
	key := strconv.FormatUint(uint64(storeId), 10)
	if j.Metadata[metadataStoreId] != key {
		return nil, apperror.New(errcode.NotFound, apperror.WithMsg("作业不存在"))
	}
	return j, nil
}

// List 分页查询作业列表（按创建时间倒序），并返回队列中当前存在的作业总数。
// jobs 表无 store_id 列，按 metadata 内存过滤当前企业作业，分页在过滤前生效
func (s *JobService) List(ctx context.Context, storeId uint32, page, limit int) ([]*job.Job, int64, error) {
	list, err := s.store.List(ctx, job.JobFilter{
		Queue:  job.DefaultQueue,
		Limit:  limit,
		Offset: (page - 1) * limit,
	})
	if err != nil {
		return nil, 0, err
	}
	stats, err := s.store.GetStats(ctx, job.DefaultQueue)
	if err != nil {
		return nil, 0, err
	}
	total := stats.Pending + stats.Running + stats.Success + stats.Failed + stats.Cancelled

	key := strconv.FormatUint(uint64(storeId), 10)
	out := make([]*job.Job, 0, len(list))
	for _, j := range list {
		if j.Metadata[metadataStoreId] == key {
			out = append(out, j)
		}
	}
	return out, total, nil
}
