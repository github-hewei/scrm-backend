package sync

import (
	"context"

	"zero-backend/internal/cli/runner"
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

// SyncStore 同步指定企业数据。实现 runner.WecomSyncService 接口
func (s *Service) SyncStore(ctx context.Context, storeId uint32, scope runner.Scope) error {
	// 构建SDK客户端
	client, err := s.clientMgr.Get(ctx, storeId)
	if err != nil {
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsgf("构建企微客户端失败 store_id=%d", storeId))
	}

	switch scope {
	case runner.ScopeGroup:
		return s.groupSvc.Sync(ctx, client, storeId)
	case runner.ScopeAll:
		// TODO: 依次同步通讯录/客户/客户群
		return s.groupSvc.Sync(ctx, client, storeId)
	default:
		// TODO: 通讯录/客户同步器接入后支持
		return apperror.New(errcode.InvalidInput, apperror.WithMsgf("暂不支持同步范围: %s", scope))
	}
}
