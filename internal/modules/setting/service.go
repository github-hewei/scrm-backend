package setting

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/241x/zero-kit/apperror"
	"github.com/241x/zero-kit/baserepo"
	"github.com/241x/zero-web/errcode"
)

// Service 设置服务
type Service struct {
	repo        *Repository
	defaultRepo *DefaultRepository
}

// NewService 创建设置服务
func NewService(repo *Repository, defaultRepo *DefaultRepository) *Service {
	return &Service{repo: repo, defaultRepo: defaultRepo}
}

// FindList 获取设置列表
func (s *Service) FindList(ctx context.Context, req *ListRequest) (*ListResult, error) {
	result := &ListResult{List: []*Setting{}, Total: 0}
	filter := &Filter{SettingKey: req.SettingKey, StoreId: req.StoreId}
	total, err := s.repo.Count(ctx, filter)
	if err != nil {
		return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("获取设置列表失败"))
	}
	if total == 0 {
		return result, nil
	}
	result.Total = total
	orders := baserepo.Orders{{Field: "id", Sort: "desc"}}
	list, err := s.repo.FindAll(ctx, filter, baserepo.NewPagination(req.Page, req.Limit), orders)
	if err != nil {
		return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("获取设置列表失败"))
	}
	result.List = list
	return result, nil
}

// Create 创建设置
func (s *Service) Create(ctx context.Context, req *CreateRequest) error {
	if err := s.checkSettingKey(ctx, req.SettingKey, req.StoreId); err != nil {
		return err
	}
	item := &Setting{SettingKey: req.SettingKey, SettingValues: req.SettingValues, Description: req.Description, StoreId: req.StoreId}
	if err := s.repo.Create(ctx, item); err != nil {
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("创建设置失败"))
	}
	return nil
}

// Update 更新设置
func (s *Service) Update(ctx context.Context, req *UpdateRequest) error {
	filter := &Filter{Id: req.ID, StoreId: req.StoreId}
	item, err := s.repo.FindOne(ctx, filter)
	if err != nil {
		if errors.Is(err, baserepo.ErrRecordNotFound) {
			return apperror.New(errcode.NotFound, apperror.WithMsg("设置不存在或无权限访问"))
		}
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("更新设置失败"))
	}
	if item.SettingKey != req.SettingKey {
		if err := s.checkSettingKey(ctx, req.SettingKey, req.StoreId); err != nil {
			return err
		}
	}
	updateData := map[string]any{"setting_key": req.SettingKey, "setting_values": req.SettingValues, "description": req.Description, "store_id": req.StoreId}
	if err := s.repo.Updates(ctx, item, updateData); err != nil {
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("更新设置失败"))
	}
	return nil
}

// Delete 删除设置
func (s *Service) Delete(ctx context.Context, req *DeleteRequest) error {
	filter := &Filter{Id: req.ID, StoreId: req.StoreId}
	item, err := s.repo.FindOne(ctx, filter)
	if err != nil {
		if errors.Is(err, baserepo.ErrRecordNotFound) {
			return apperror.New(errcode.NotFound)
		}
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("删除设置失败"))
	}
	if err := s.repo.Delete(ctx, item.ID); err != nil {
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("删除设置失败"))
	}
	return nil
}

// checkSettingKey 检查设置key
func (s *Service) checkSettingKey(ctx context.Context, key string, storeId uint32) error {
	filter := Filter{SettingKey: key, StoreId: storeId}
	_, err := s.repo.FindOne(ctx, filter)
	if err != nil {
		if errors.Is(err, baserepo.ErrRecordNotFound) {
			return nil
		}
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("检查设置key失败"))
	}
	return apperror.New(errcode.Conflict, apperror.WithMsg("设置key已存在"))
}

// GetSettingValue 获取并解析设置项
func (s *Service) GetSettingValue(ctx context.Context, key string, out any) error {
	filter := Filter{SettingKey: key}
	setting, err := s.repo.FindOne(ctx, filter)
	var settingValues string
	if err != nil {
		if !errors.Is(err, baserepo.ErrRecordNotFound) {
			return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("获取设置项失败"))
		}
		defaultSetting, err := s.defaultRepo.FindOne(ctx, &DefaultFilter{SettingKey: key}, baserepo.WithScopes(nil))
		if err != nil {
			if errors.Is(err, baserepo.ErrRecordNotFound) {
				return apperror.New(errcode.NotFound, apperror.WithMsg("设置项不存在"))
			}
			return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("获取设置项失败"))
		}
		settingValues = defaultSetting.SettingValues
	} else {
		settingValues = setting.SettingValues
	}
	if err := json.Unmarshal([]byte(settingValues), out); err != nil {
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("解析设置项失败"))
	}
	return nil
}

// FormConfigs 获取表单配置
func (s *Service) FormConfigs(ctx context.Context, req *FormConfigsRequest) ([]FormGroup, error) {
	configs := FormConfigs()
	if req.OnlyPlatform {
		filtered := make([]FormGroup, 0)
		for _, config := range configs {
			if !config.OnlyPlatform {
				filtered = append(filtered, config)
			}
		}
		return filtered, nil
	}
	return configs, nil
}

// DefaultService 默认设置服务
type DefaultService struct {
	repo *DefaultRepository
}

// NewDefaultService 创建默认设置服务
func NewDefaultService(repo *DefaultRepository) *DefaultService {
	return &DefaultService{repo: repo}
}

// FindList 获取默认设置列表
func (s *DefaultService) FindList(ctx context.Context, req *DefaultListRequest) (*ListResult, error) {
	result := &ListResult{List: []*SettingDefault{}, Total: 0}
	filter := &DefaultFilter{SettingKey: req.SettingKey}
	total, err := s.repo.Count(ctx, filter, baserepo.WithScopes(nil))
	if err != nil {
		return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("获取默认设置列表失败"))
	}
	if total == 0 {
		return result, nil
	}
	result.Total = total
	orders := baserepo.Orders{{Field: "id", Sort: "desc"}}
	list, err := s.repo.FindAll(ctx, filter, baserepo.NewPagination(req.Page, req.Limit), orders, baserepo.WithScopes(nil))
	if err != nil {
		return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("获取默认设置列表失败"))
	}
	result.List = list
	return result, nil
}

// Create 创建默认设置
func (s *DefaultService) Create(ctx context.Context, req *DefaultCreateRequest) error {
	if err := s.checkSettingKey(ctx, req.SettingKey); err != nil {
		return err
	}
	item := &SettingDefault{SettingKey: req.SettingKey, SettingValues: req.SettingValues, Description: req.Description}
	if err := s.repo.Create(ctx, item); err != nil {
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("创建默认设置失败"))
	}
	return nil
}

// Update 更新默认设置
func (s *DefaultService) Update(ctx context.Context, req *DefaultUpdateRequest) error {
	item, err := s.repo.FindOne(ctx, req.ID)
	if err != nil {
		if errors.Is(err, baserepo.ErrRecordNotFound) {
			return apperror.New(errcode.NotFound, apperror.WithMsg("默认设置不存在"))
		}
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("更新默认设置失败"))
	}
	if item.SettingKey != req.SettingKey {
		if err := s.checkSettingKey(ctx, req.SettingKey); err != nil {
			return err
		}
	}
	updateData := map[string]any{"setting_key": req.SettingKey, "setting_values": req.SettingValues, "description": req.Description}
	if err := s.repo.Updates(ctx, item, updateData); err != nil {
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("更新默认设置失败"))
	}
	return nil
}

// Delete 删除默认设置
func (s *DefaultService) Delete(ctx context.Context, req *DefaultDeleteRequest) error {
	if _, err := s.repo.FindOne(ctx, req.ID); err != nil {
		if errors.Is(err, baserepo.ErrRecordNotFound) {
			return apperror.New(errcode.NotFound)
		}
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("删除默认设置失败"))
	}
	return s.repo.Delete(ctx, req.ID)
}

// checkSettingKey 检查默认设置key
func (s *DefaultService) checkSettingKey(ctx context.Context, key string) error {
	filter := &DefaultFilter{SettingKey: key}
	_, err := s.repo.FindOne(ctx, filter, baserepo.WithScopes(nil))
	if err != nil {
		if errors.Is(err, baserepo.ErrRecordNotFound) {
			return nil
		}
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("检查默认设置key失败"))
	}
	return apperror.New(errcode.Conflict, apperror.WithMsg("默认设置key已存在"))
}
