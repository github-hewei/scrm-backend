package sync

import (
	"context"
	"errors"

	"zero-backend/internal/modules/wecom/config"

	"github.com/241x/zero-kit/apperror"
	"github.com/241x/zero-kit/baserepo"
	"github.com/241x/zero-third/wecom"
	"github.com/241x/zero-web/errcode"
	"gorm.io/gorm"
)

// ClientManager 企微SDK客户端管理器：按企业构建 wecom.Client
type ClientManager struct {
	configRepo *config.WecomConfigRepository
	appRepo    *config.WecomAppRepository
	cache      wecom.TokenCache
}

// NewClientManager 创建客户端管理器
func NewClientManager(configRepo *config.WecomConfigRepository, appRepo *config.WecomAppRepository, cache wecom.TokenCache) *ClientManager {
	return &ClientManager{configRepo: configRepo, appRepo: appRepo, cache: cache}
}

// Get 获取企业的SDK客户端。使用最新配置的自建应用凭据（权限全开覆盖全部数据域）。
// agentID 仅发消息时需要，通讯录/客户同步场景传 0 即可
func (m *ClientManager) Get(ctx context.Context, storeId uint32) (*wecom.Client, error) {
	wecomConfig, err := m.configRepo.FindOne(ctx, &config.WecomConfigFilter{StoreId: storeId})
	if err != nil {
		return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsgf("获取企业接入配置失败 store_id=%d", storeId))
	}

	app, err := m.resolveApp(ctx, storeId)
	if err != nil {
		return nil, err
	}

	opts := []wecom.ClientOption{}
	if wecomConfig.ApiBaseUrl != "" {
		opts = append(opts, wecom.WithBaseURL(wecomConfig.ApiBaseUrl))
	}
	client, err := wecom.NewClient(wecom.CorpSecret{CorpID: wecomConfig.CorpId, Secret: app.Secret}, 0, m.cache, opts...)
	if err != nil {
		return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("创建企微客户端失败"))
	}
	return client, nil
}

// resolveApp 获取企业最新配置的自建应用凭据，未配置则报错
func (m *ClientManager) resolveApp(ctx context.Context, storeId uint32) (*config.WecomApp, error) {
	app, err := m.appRepo.FindOne(ctx, &config.WecomAppFilter{StoreId: storeId, AppType: int8(config.AppTypeSelfBuilt)},
		baserepo.WithScopes(func(db *gorm.DB) *gorm.DB { return db.Order("id desc") }))
	if err != nil {
		if errors.Is(err, baserepo.ErrRecordNotFound) {
			return nil, apperror.New(errcode.NotFound, apperror.WithMsgf("企业未配置自建应用 store_id=%d", storeId))
		}
		return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsgf("查询自建应用失败 store_id=%d", storeId))
	}
	if app.Secret == "" {
		return nil, apperror.New(errcode.InvalidInput, apperror.WithMsgf("自建应用密钥未配置 store_id=%d", storeId))
	}
	return app, nil
}
