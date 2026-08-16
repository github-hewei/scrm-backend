package async

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/241x/zero-kit/apperror"
	"github.com/241x/zero-kit/baserepo"
	"github.com/241x/zero-kit/job"
	"github.com/241x/zero-web/errcode"
)

// TaskService 通用异步任务服务：任务提交、执行状态同步与租户查询
type TaskService struct {
	repo        *AsyncTaskRepository
	store       *job.SQLStore
	maxAttempts int
}

// NewTaskService 创建异步任务服务
func NewTaskService(repo *AsyncTaskRepository, store *job.SQLStore) *TaskService {
	return &TaskService{repo: repo, store: store, maxAttempts: 3}
}

// Submit 提交异步任务：按 task_type 查注册规则 → 业务校验 → 连点去重 →
// 创建系统调度作业并登记任务记录，返回任务ID（租户业务主键，用于详情/列表查询）。
// 提交成功后由 worker 进程异步执行
func (s *TaskService) Submit(ctx context.Context, storeId uint32, taskType string, payload []byte) (uint64, error) {
	rule, ok := lookupSubmitHandler(taskType)
	if !ok {
		return 0, apperror.New(errcode.InvalidInput, apperror.WithMsgf("不支持的任务类型: %s", taskType))
	}
	if storeId == 0 {
		return 0, apperror.New(errcode.InvalidInput, apperror.WithMsg("企业标识缺失"))
	}
	if rule.Validate != nil {
		if err := rule.Validate(ctx, storeId, payload); err != nil {
			return 0, err
		}
	}

	// 连点去重：同类型同去重键的 pending/running 作业存在时拒绝重复提交。
	// 作业按创建时间倒序，进行中作业必在最新一批内，Limit 取 100 已覆盖实际并发
	if rule.DedupKey != nil {
		key := rule.DedupKey(payload)
		if key != "" {
			existing, err := s.store.List(ctx, job.JobFilter{Queue: job.DefaultQueue, Type: taskType, Limit: 100})
			if err != nil {
				return 0, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("查询进行中作业失败"))
			}
			for _, j := range existing {
				if j.Status != job.StatusPending && j.Status != job.StatusRunning {
					continue
				}
				if rule.DedupKey(j.Payload) == key {
					return 0, apperror.New(errcode.Conflict, apperror.WithMsgf("已有相同任务在执行 job_id=%s", j.ID))
				}
			}
		}
	}

	title := ""
	if rule.Title != nil {
		title = rule.Title(payload)
	}
	maxAttempts := s.maxAttempts
	if rule.MaxAttempts > 0 {
		maxAttempts = rule.MaxAttempts
	}
	j := job.NewJob(taskType, payload).WithMaxAttempts(maxAttempts)
	if err := s.store.Save(ctx, j); err != nil {
		return 0, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("提交任务失败"))
	}
	task, err := s.Create(ctx, storeId, j.ID, taskType, title)
	if err != nil {
		// 登记失败时回滚已入队作业，避免"作业已调度但无任务记录"的不一致
		if delErr := s.store.Delete(ctx, j.ID); delErr != nil {
			err = errors.Join(err, fmt.Errorf("回滚作业失败: %w", delErr))
		}
		return 0, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("登记任务记录失败"))
	}
	return task.ID, nil
}

// Create 创建任务记录（提交异步作业后调用，与 jobs 表 1:1 关联），返回创建的任务
func (s *TaskService) Create(ctx context.Context, storeId uint32, jobId, taskType, title string) (*AsyncTask, error) {
	if jobId == "" || taskType == "" {
		return nil, errors.New("async: job_id and task_type are required")
	}
	now := uint32(time.Now().Unix())
	task := &AsyncTask{
		StoreId:   storeId,
		JobId:     jobId,
		TaskType:  taskType,
		Title:     title,
		Status:    TaskStatusPending,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.repo.Create(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}

// markRunning 按已加载任务标记执行中
func (s *TaskService) markRunning(ctx context.Context, task *AsyncTask) error {
	return s.repo.Updates(ctx, task, map[string]any{
		"status":     string(TaskStatusRunning),
		"updated_at": uint32(time.Now().Unix()),
	})
}

// markProgress 按已加载任务更新进度
func (s *TaskService) markProgress(ctx context.Context, task *AsyncTask, progress int) error {
	if progress < 0 {
		progress = 0
	}
	if progress > 100 {
		progress = 100
	}
	return s.repo.Updates(ctx, task, map[string]any{
		"progress":   progress,
		"updated_at": uint32(time.Now().Unix()),
	})
}

// markFinish 按已加载任务写入终态与结果摘要
func (s *TaskService) markFinish(ctx context.Context, task *AsyncTask, runErr error, result []byte) error {
	updates := map[string]any{"updated_at": uint32(time.Now().Unix())}
	if runErr != nil {
		updates["status"] = string(TaskStatusFailed)
		updates["error"] = runErr.Error()
	} else {
		updates["status"] = string(TaskStatusSuccess)
		updates["progress"] = 100
	}
	if len(result) > 0 {
		updates["result"] = result
	}
	return s.repo.Updates(ctx, task, updates)
}

// markCancelled 按已加载任务标记取消
func (s *TaskService) markCancelled(ctx context.Context, task *AsyncTask) error {
	return s.repo.Updates(ctx, task, map[string]any{
		"status":     string(TaskStatusCancelled),
		"updated_at": uint32(time.Now().Unix()),
	})
}

// ListByStore 分页查询企业任务（按创建倒序），返回当前企业全部历史任务数
func (s *TaskService) ListByStore(ctx context.Context, storeId uint32, taskType string, page, limit int) ([]*AsyncTask, int64, error) {
	filter := &AsyncTaskFilter{StoreId: storeId}
	if taskType != "" {
		filter.TaskType = taskType
	}
	total, err := s.repo.Count(ctx, filter)
	if err != nil {
		return nil, 0, err
	}
	list, err := s.repo.FindAll(ctx, filter, baserepo.NewPagination(page, limit), baserepo.Orders{{Field: "id", Sort: "DESC"}})
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// GetByStore 查询企业任务详情（校验归属，跨企业返回 NotFound）
func (s *TaskService) GetByStore(ctx context.Context, storeId uint32, id uint64) (*AsyncTask, error) {
	task, err := s.repo.FindOne(ctx, &AsyncTaskFilter{StoreId: storeId, Id: id})
	if err != nil {
		if errors.Is(err, baserepo.ErrRecordNotFound) {
			return nil, apperror.New(errcode.NotFound, apperror.WithMsg("任务不存在"))
		}
		return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("查询任务失败"))
	}
	return task, nil
}

// findByJob 按 job_id 查询任务，无记录返回 (nil, nil)
func (s *TaskService) findByJob(ctx context.Context, jobId string) (*AsyncTask, error) {
	task, err := s.repo.FindOne(ctx, &AsyncTaskFilter{JobId: jobId})
	if err != nil {
		if errors.Is(err, baserepo.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("查询任务失败"))
	}
	return task, nil
}

// Recorder 作业执行记录器：handler 内同步任务状态，消除各业务 handler 重复代码。
// 任务按 job_id 懒加载并缓存，后续更新按主键执行，避免每次操作重复查询。
// 所有 DB 操作使用 WithoutCancel 派生 ctx：作业被取消/超时后终态仍能落库，
// 避免任务记录因 handler ctx 已取消而永久停留在 running
type Recorder struct {
	svc    *TaskService
	jobID  string
	task   *AsyncTask
	loaded bool
}

// NewRecorder 创建作业执行记录器
func NewRecorder(svc *TaskService, jobID string) *Recorder {
	return &Recorder{svc: svc, jobID: jobID}
}

// dbCtx 返回与作业取消无关的上下文（保留 Values，切掉取消信号）
func (r *Recorder) dbCtx(ctx context.Context) context.Context {
	return context.WithoutCancel(ctx)
}

// load 懒加载任务记录；无对应记录（历史任务）时返回 (nil, nil)，后续操作均为空操作
func (r *Recorder) load(ctx context.Context) (*AsyncTask, error) {
	if r.loaded {
		return r.task, nil
	}
	r.loaded = true
	task, err := r.svc.findByJob(r.dbCtx(ctx), r.jobID)
	if err != nil {
		return nil, err
	}
	r.task = task
	return task, nil
}

// Start 标记任务执行中
func (r *Recorder) Start(ctx context.Context) error {
	task, err := r.load(ctx)
	if err != nil || task == nil {
		return err
	}
	return r.svc.markRunning(r.dbCtx(ctx), task)
}

// Progress 更新任务进度
func (r *Recorder) Progress(ctx context.Context, progress int) error {
	task, err := r.load(ctx)
	if err != nil || task == nil {
		return err
	}
	return r.svc.markProgress(r.dbCtx(ctx), task, progress)
}

// Cancel 标记任务取消（作业被用户取消时调用，语义区别于失败）
func (r *Recorder) Cancel(ctx context.Context) error {
	task, err := r.load(ctx)
	if err != nil || task == nil {
		return err
	}
	return r.svc.markCancelled(r.dbCtx(ctx), task)
}

// Finish 标记任务终态并写入结果摘要
func (r *Recorder) Finish(ctx context.Context, runErr error, result []byte) error {
	task, err := r.load(ctx)
	if err != nil || task == nil {
		return err
	}
	return r.svc.markFinish(r.dbCtx(ctx), task, runErr, result)
}
