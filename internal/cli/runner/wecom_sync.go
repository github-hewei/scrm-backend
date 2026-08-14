package runner

import (
	"context"

	wecomsync "zero-backend/internal/modules/wecom/sync"

	"github.com/241x/zero-kit/logger"
)

// WecomSyncService 企微同步服务接口，由 wecom/sync 包实现。
// runner 不直接依赖具体实现，便于后续调度器复用同一内核
type WecomSyncService interface {
	// LoadStoreIds 加载待同步企业ID列表（storeId=0 时返回全部已接入企业）
	LoadStoreIds(ctx context.Context, storeId uint32) ([]uint32, error)
	// SyncStore 同步指定企业数据，scope 指定数据域
	SyncStore(ctx context.Context, storeId uint32, scope wecomsync.Scope) error
	// SyncStores 批量同步企业：逐企业调用，每完成一个回调进度，聚合错误返回
	SyncStores(ctx context.Context, storeIds []uint32, scope wecomsync.Scope, onProgress func(storeId uint32, completed, total int, err error)) error
}

// WecomSyncRunner 企微同步执行器：解析企业列表并调用同步服务
type WecomSyncRunner struct {
	log logger.Logger
	svc WecomSyncService
}

// NewWecomSyncRunner 创建企微同步执行器
func NewWecomSyncRunner(log logger.Logger, svc WecomSyncService) *WecomSyncRunner {
	return &WecomSyncRunner{log: log, svc: svc}
}

// Run 执行同步。storeId=0 时同步全部已接入企业；任一企业失败返回聚合错误（退出码非0）
func (r *WecomSyncRunner) Run(ctx context.Context, storeId uint32, scope string) error {
	parsedScope, err := wecomsync.ParseScope(scope)
	if err != nil {
		return err
	}

	storeIds, err := r.svc.LoadStoreIds(ctx, storeId)
	if err != nil {
		return err
	}
	if len(storeIds) == 0 {
		r.log.Info("没有需要同步的企业")
		return nil
	}

	r.log.Info("开始同步企业", "count", len(storeIds), "scope", parsedScope)
	if err := r.svc.SyncStores(ctx, storeIds, parsedScope, nil); err != nil {
		r.log.Err(err, "同步失败")
		return err
	}
	r.log.Info("企业同步完成", "count", len(storeIds), "scope", parsedScope)
	return nil
}
