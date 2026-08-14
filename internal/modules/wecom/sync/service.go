package sync

import (
	"context"
	"errors"
	"fmt"

	"zero-backend/internal/modules/wecom/config"
	"zero-backend/internal/modules/wecom/group"

	"github.com/241x/zero-kit/apperror"
	"github.com/241x/zero-web/errcode"
)

// Service 企微数据同步服务：编排各子包同步器
type Service struct {
	configRepo *config.WecomConfigRepository
	clientMgr  *ClientManager
	groupSvc   *group.GroupSyncer
}

// NewService 创建同步服务
func NewService(configRepo *config.WecomConfigRepository, clientMgr *ClientManager, groupSvc *group.GroupSyncer) *Service {
	return &Service{configRepo: configRepo, clientMgr: clientMgr, groupSvc: groupSvc}
}

// ListActiveStoreIds 获取全部已接入企业ID列表（wecom_config.status=1）
func (s *Service) ListActiveStoreIds(ctx context.Context) ([]uint32, error) {
	list, err := s.configRepo.FindAll(ctx, &config.WecomConfigFilter{Status: 1}, nil, nil)
	if err != nil {
		return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("获取已接入企业列表失败"))
	}
	storeIds := make([]uint32, 0, len(list))
	for _, item := range list {
		storeIds = append(storeIds, item.StoreId)
	}
	return storeIds, nil
}

// LoadStoreIds 加载待同步企业ID列表：storeId=0 时返回全部已接入(status=1)企业
func (s *Service) LoadStoreIds(ctx context.Context, storeId uint32) ([]uint32, error) {
	if storeId != 0 {
		return []uint32{storeId}, nil
	}
	return s.ListActiveStoreIds(ctx)
}

// SyncStores 批量同步多个企业：循环调 SyncStore，每完成一个企业回调 onProgress(storeId, completed, total, err)，
// 全部完成后聚合错误返回（单个企业失败不阻断其它企业）
func (s *Service) SyncStores(ctx context.Context, storeIds []uint32, scope Scope, onProgress func(storeId uint32, completed, total int, err error)) error {
	errs := make([]error, 0, len(storeIds))
	for i, sid := range storeIds {
		syncErr := s.SyncStore(ctx, sid, scope)
		if syncErr != nil {
			errs = append(errs, fmt.Errorf("store_id=%d: %w", sid, syncErr))
		}
		if onProgress != nil {
			onProgress(sid, i+1, len(storeIds), syncErr)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("同步失败 %d/%d 个企业: %w", len(errs), len(storeIds), errors.Join(errs...))
	}
	return nil
}

// SyncStore 同步指定企业数据
func (s *Service) SyncStore(ctx context.Context, storeId uint32, scope Scope) error {
	// 构建SDK客户端
	client, err := s.clientMgr.Get(ctx, storeId)
	if err != nil {
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsgf("构建企微客户端失败 store_id=%d", storeId))
	}

	// 未接入执行器的范围直接拒绝，避免静默退化（如 ScopeAll 只同步部分数据）
	if !IsSupported(scope) {
		return apperror.New(errcode.InvalidInput, apperror.WithMsgf("暂不支持同步范围: %s", scope))
	}

	switch scope {
	case ScopeGroup:
		return s.groupSvc.Sync(ctx, client, storeId)
	default:
		return apperror.New(errcode.Internal, apperror.WithMsgf("同步范围未接入执行器: %s", scope))
	}
}
