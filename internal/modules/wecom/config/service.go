package config

import (
	"context"
	"errors"
	"strings"

	"github.com/241x/zero-kit/apperror"
	"github.com/241x/zero-kit/baserepo"
	"github.com/241x/zero-kit/helper"
	"github.com/241x/zero-web/errcode"
)

// SettingProvider 系统设置读取接口，由外部注入实现
// 用于获取站点域名拼接企微回调地址
type SettingProvider interface {
	GetSettingValue(ctx context.Context, key string, storeId uint32, target any) error
}

// Service 企业微信配置服务
type Service struct {
	configRepo *WecomConfigRepository
	appRepo    *WecomAppRepository
	settings   SettingProvider
}

// NewService 创建企业微信配置服务
func NewService(configRepo *WecomConfigRepository, appRepo *WecomAppRepository, settings SettingProvider) *Service {
	return &Service{configRepo: configRepo, appRepo: appRepo, settings: settings}
}

// GetConfig 获取企业微信配置
func (s *Service) GetConfig(ctx context.Context, storeId uint32) (*ConfigInfo, error) {
	item, err := s.configRepo.FindOne(ctx, &WecomConfigFilter{StoreId: storeId})
	if err != nil {
		if errors.Is(err, baserepo.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("获取企业微信配置失败"))
	}
	return &ConfigInfo{CorpId: item.CorpId, CorpName: item.CorpName, ApiBaseUrl: item.ApiBaseUrl, Status: item.Status}, nil
}

// SaveConfig 保存企业微信配置。首次创建时置未接入状态，更新时保留原状态
func (s *Service) SaveConfig(ctx context.Context, storeId uint32, req *ConfigSaveRequest) error {
	item, err := s.configRepo.FindOne(ctx, &WecomConfigFilter{StoreId: storeId})
	if err != nil && !errors.Is(err, baserepo.ErrRecordNotFound) {
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("保存企业微信配置失败"))
	}

	if errors.Is(err, baserepo.ErrRecordNotFound) {
		item = &WecomConfig{StoreId: storeId, CorpId: req.CorpId, CorpName: req.CorpName, ApiBaseUrl: req.ApiBaseUrl, Status: 0}
		if err := s.configRepo.Create(ctx, item); err != nil {
			return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("保存企业微信配置失败"))
		}
		return nil
	}

	updateData := map[string]any{
		"corp_id":      req.CorpId,
		"corp_name":    req.CorpName,
		"api_base_url": req.ApiBaseUrl,
	}
	if err := s.configRepo.Updates(ctx, item, updateData); err != nil {
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("保存企业微信配置失败"))
	}
	return nil
}

// GetCallbackUrl 生成回调地址。一次性生成随机token，不入库，保存时由前端回传
func (s *Service) GetCallbackUrl(ctx context.Context, storeId uint32) (*CallbackUrlResponse, error) {
	token := helper.RandomString(16)
	return &CallbackUrlResponse{
		CallbackToken: token,
		CallbackUrl:   s.buildCallbackUrl(ctx, token, storeId),
	}, nil
}

// ListApps 获取应用凭据列表
func (s *Service) ListApps(ctx context.Context, storeId uint32) (*AppListResponse, error) {
	list, err := s.appRepo.FindAll(ctx, &WecomAppFilter{StoreId: storeId}, nil, nil)
	if err != nil {
		return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("获取应用列表失败"))
	}
	result := &AppListResponse{List: make([]*AppInfo, 0, len(list))}
	for _, item := range list {
		result.List = append(result.List, &AppInfo{
			ID:            item.ID,
			AppType:       AppType(item.AppType),
			AppTypeText:   AppType(item.AppType).Name(),
			AppName:       item.AppName,
			AgentId:       item.AgentId,
			SecretSet:     item.Secret != "",
			TokenSet:      item.Token != "",
			CallbackToken: item.CallbackToken,
			CallbackUrl:   s.buildCallbackUrl(ctx, item.CallbackToken, storeId),
			Status:        item.Status,
		})
	}
	return result, nil
}

// SaveApp 保存应用凭据。自建应用以(storeId,agentId)为键，客户联系/通讯录以(storeId,appType)为键。
// 首次创建需传入callback_token；更新时不修改callback_token(企微回调地址已按其配置，不可变)
func (s *Service) SaveApp(ctx context.Context, storeId uint32, req *AppSaveRequest) (*AppSaveResponse, error) {
	if req.AppType == AppTypeSelfBuilt && req.AgentId == 0 {
		return nil, apperror.New(errcode.InvalidInput, apperror.WithMsg("自建应用必须填写AgentID"))
	}

	item, err := s.findAppForSave(ctx, storeId, req)
	if err != nil && !errors.Is(err, baserepo.ErrRecordNotFound) {
		return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("保存应用失败"))
	}

	if errors.Is(err, baserepo.ErrRecordNotFound) {
		if req.CallbackToken == "" {
			return nil, apperror.New(errcode.InvalidInput, apperror.WithMsg("请先获取回调地址"))
		}
		item = &WecomApp{
			StoreId: storeId, AppType: int8(req.AppType), AppName: req.AppName,
			CallbackToken: req.CallbackToken, AgentId: req.AgentId,
			Secret: req.Secret, Token: req.Token,
			EncodingAesKey: req.EncodingAesKey, Status: 1,
		}
		if err := s.appRepo.Create(ctx, item); err != nil {
			return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("保存应用失败"))
		}
	} else {
		updateData := map[string]any{
			"app_name":         req.AppName,
			"agent_id":         req.AgentId,
			"secret":           req.Secret,
			"token":            req.Token,
			"encoding_aes_key": req.EncodingAesKey,
		}
		if err := s.appRepo.Updates(ctx, item, updateData); err != nil {
			return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("保存应用失败"))
		}
	}

	return &AppSaveResponse{
		ID:            item.ID,
		CallbackToken: item.CallbackToken,
		CallbackUrl:   s.buildCallbackUrl(ctx, item.CallbackToken, storeId),
		SecretSet:     item.Secret != "",
		TokenSet:      item.Token != "",
	}, nil
}

// buildCallbackUrl 拼接企微回调地址 {site.domain}/wecom/callback/{token}（domain含协议前缀）
func (s *Service) buildCallbackUrl(ctx context.Context, callbackToken string, storeId uint32) string {
	if s.settings == nil || callbackToken == "" {
		return ""
	}
	var site struct {
		Domain string `json:"domain"`
	}
	if err := s.settings.GetSettingValue(ctx, "site", storeId, &site); err != nil || site.Domain == "" {
		return ""
	}
	return strings.TrimRight(site.Domain, "/") + "/wecom/callback/" + callbackToken
}

// findAppForSave 按类型定位待更新应用
func (s *Service) findAppForSave(ctx context.Context, storeId uint32, req *AppSaveRequest) (*WecomApp, error) {
	if req.AppType == AppTypeSelfBuilt {
		return s.appRepo.FindOne(ctx, &WecomAppFilter{StoreId: storeId, AgentId: req.AgentId})
	}
	return s.appRepo.FindOne(ctx, &WecomAppFilter{StoreId: storeId, AppType: int8(req.AppType)})
}

// DeleteApp 删除应用凭据
func (s *Service) DeleteApp(ctx context.Context, storeId, id uint32) error {
	item, err := s.appRepo.FindOne(ctx, &WecomAppFilter{Id: id, StoreId: storeId})
	if err != nil {
		if errors.Is(err, baserepo.ErrRecordNotFound) {
			return apperror.New(errcode.NotFound, apperror.WithMsg("应用不存在或无权限访问"))
		}
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("删除应用失败"))
	}
	if err := s.appRepo.Delete(ctx, item.ID); err != nil {
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("删除应用失败"))
	}
	return nil
}
