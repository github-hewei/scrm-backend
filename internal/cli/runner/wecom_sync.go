package runner

import (
	"context"
	"fmt"

	"github.com/241x/zero-kit/logger"
)

// Scope 同步范围
type Scope string

// 同步范围定义
const (
	ScopeAll     Scope = "all"     // 全部（通讯录+客户+客户群）
	ScopeDept    Scope = "dept"    // 通讯录（部门+成员）
	ScopeContact Scope = "contact" // 外部联系人（客户）
	ScopeGroup   Scope = "group"   // 客户群
)

// ParseScope 解析同步范围参数，返回是否合法
func ParseScope(s string) (Scope, error) {
	switch Scope(s) {
	case ScopeAll, ScopeDept, ScopeContact, ScopeGroup:
		return Scope(s), nil
	default:
		return "", fmt.Errorf("无效的同步范围: %s (可选 all|dept|contact|group)", s)
	}
}

// WecomSyncService 企微同步服务接口，由 wecom/sync 包实现
// runner 不直接依赖具体实现，便于后续调度器复用同一内核
type WecomSyncService interface {
	// ListActiveStoreIds 获取全部已接入企业ID列表
	ListActiveStoreIds(ctx context.Context) ([]uint32, error)
	// SyncStore 同步指定企业数据，scope 指定数据域
	SyncStore(ctx context.Context, storeId uint32, scope Scope) error
}

// WecomSyncRunner 企微同步执行器：解析企业列表并逐企业调用同步服务
type WecomSyncRunner struct {
	log logger.Logger
	svc WecomSyncService
}

// NewWecomSyncRunner 创建企微同步执行器
func NewWecomSyncRunner(log logger.Logger, svc WecomSyncService) *WecomSyncRunner {
	return &WecomSyncRunner{log: log, svc: svc}
}

// Run 执行同步。storeId=0 时同步全部已接入企业
func (r *WecomSyncRunner) Run(ctx context.Context, storeId uint32, scope string) error {
	parsedScope, err := ParseScope(scope)
	if err != nil {
		return err
	}

	storeIds, err := r.loadStoreIds(ctx, storeId)
	if err != nil {
		return err
	}
	if len(storeIds) == 0 {
		r.log.Info("没有需要同步的企业")
		return nil
	}

	for _, sid := range storeIds {
		r.log.Info("开始同步企业", "store_id", sid, "scope", parsedScope)
		if err := r.svc.SyncStore(ctx, sid, parsedScope); err != nil {
			r.log.Err(err, "同步企业失败", "store_id", sid)
			continue // 单个企业失败不阻断其它企业
		}
		r.log.Info("企业同步完成", "store_id", sid, "scope", parsedScope)
	}
	return nil
}

// loadStoreIds 加载待同步企业ID列表。
// storeId 指定时只同步该企业；否则查询全部已接入(status=1)企业
func (r *WecomSyncRunner) loadStoreIds(ctx context.Context, storeId uint32) ([]uint32, error) {
	if storeId != 0 {
		return []uint32{storeId}, nil
	}
	storeIds, err := r.svc.ListActiveStoreIds(ctx)
	if err != nil {
		return nil, err
	}
	return storeIds, nil
}
