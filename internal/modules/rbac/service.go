package rbac

import (
	"context"
	"errors"

	"github.com/241x/zero-kit/apperror"
	"github.com/241x/zero-kit/baserepo"
	"github.com/241x/zero-kit/helper"
	"github.com/241x/zero-web/errcode"
	"gorm.io/gorm"
)

// RbacUserService 用户服务
type RbacUserService struct {
	db           *gorm.DB
	repo         *RbacUserRepository
	userRoleRepo *RbacUserRoleRepository
	roleRepo     *RbacRoleRepository
}

// NewRbacUserService 创建用户服务
func NewRbacUserService(db *gorm.DB, repo *RbacUserRepository, userRoleRepo *RbacUserRoleRepository, roleRepo *RbacRoleRepository) *RbacUserService {
	return &RbacUserService{db: db, repo: repo, userRoleRepo: userRoleRepo, roleRepo: roleRepo}
}

// FindList 获取用户列表
func (s *RbacUserService) FindList(ctx context.Context, req *RbacUserListRequest) (*ListResult, error) {
	result := &ListResult{List: []*RbacUser{}, Total: 0}

	filter := &RbacUserFilterField{
		Username: req.Username,
		RealName: req.RealName,
		StoreId:  req.StoreId,
		IsSuper:  req.IsSuper,
	}

	pagination := baserepo.NewPagination(req.Page, req.Limit)
	orders := baserepo.Orders{
		{Field: "sort", Sort: "asc"},
		{Field: "id", Sort: "desc"},
	}

	total, err := s.repo.Count(ctx, filter, baserepo.WithScopes(nil))
	if err != nil {
		return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("获取用户列表失败"))
	}
	if total == 0 {
		return result, nil
	}
	result.Total = total

	list, err := s.repo.FindAll(ctx, filter, pagination, orders,
		baserepo.WithScopes(nil),
		baserepo.WithPreloads("RbacUserRole.RbacRole"),
	)
	if err != nil {
		return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("获取用户列表失败"))
	}
	result.List = list
	return result, nil
}

// Create 创建用户
func (s *RbacUserService) Create(ctx context.Context, req *RbacUserCreateRequest) error {
	if req.IsSuper == 1 && req.StoreId == 0 {
		return apperror.New(errcode.InvalidInput, apperror.WithMsg("创建超管用户必须指定所属企业"))
	}
	if err := s.checkUsername(ctx, req.Username, req.StoreId); err != nil {
		return err
	}

	hashedPassword, err := helper.HashPassword(req.Password)
	if err != nil {
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("密码加密失败"))
	}

	item := &RbacUser{
		Username: req.Username,
		Password: hashedPassword,
		RealName: req.RealName,
		IsSuper:  req.IsSuper,
		StoreId:  req.StoreId,
		Sort:     req.Sort,
	}

	if err := s.repo.Create(ctx, item); err != nil {
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("创建用户失败"))
	}
	return nil
}

// Update 更新用户
func (s *RbacUserService) Update(ctx context.Context, req *RbacUserUpdateRequest) error {
	item, err := s.repo.FindOne(ctx, req.ID, baserepo.WithScopes(nil))
	if err != nil {
		if errors.Is(err, baserepo.ErrRecordNotFound) {
			return apperror.New(errcode.NotFound, apperror.WithMsg("用户不存在"))
		}
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("更新用户失败"))
	}
	if item.StoreId != req.StoreId {
		return apperror.New(errcode.NotFound, apperror.WithMsg("用户不存在"))
	}
	if item.Username != req.Username {
		if err := s.checkUsername(ctx, req.Username, req.StoreId); err != nil {
			return err
		}
	}

	updateData := map[string]any{
		"username":  req.Username,
		"real_name": req.RealName,
		"store_id":  req.StoreId,
		"is_super":  req.IsSuper,
		"sort":      req.Sort,
	}

	if err := s.repo.Updates(ctx, item, updateData); err != nil {
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("更新用户失败"))
	}
	return nil
}

// Delete 删除用户
func (s *RbacUserService) Delete(ctx context.Context, req *RbacUserDeleteRequest) error {
	item, err := s.repo.FindOne(ctx, req.ID, baserepo.WithScopes(nil))
	if err != nil {
		if errors.Is(err, baserepo.ErrRecordNotFound) {
			return apperror.New(errcode.NotFound)
		}
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("删除用户失败"))
	}
	if item.StoreId != req.StoreId {
		return apperror.New(errcode.NotFound)
	}

	if err := s.repo.Delete(ctx, item.ID); err != nil {
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("删除用户失败"))
	}
	return nil
}

// SetRoles 设置用户角色
func (s *RbacUserService) SetRoles(ctx context.Context, req *RbacUserRoleSetRequest) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		user, err := s.repo.FindOne(ctx, req.UserID, baserepo.WithDB[*baserepo.QueryConfig](tx), baserepo.WithScopes(nil))
		if err != nil {
			if errors.Is(err, baserepo.ErrRecordNotFound) {
				return apperror.New(errcode.NotFound, apperror.WithMsg("用户不存在"))
			}
			return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("设置用户角色失败"))
		}
		if user.StoreId != req.StoreId {
			return apperror.New(errcode.NotFound, apperror.WithMsg("用户不存在"))
		}

		for _, roleID := range req.RoleIDs {
			role, err := s.roleRepo.FindOne(ctx, roleID, baserepo.WithDB[*baserepo.QueryConfig](tx), baserepo.WithScopes(nil))
			if err != nil {
				if errors.Is(err, baserepo.ErrRecordNotFound) {
					return apperror.New(errcode.NotFound, apperror.WithMsg("角色不存在"))
				}
				return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("设置用户角色失败"))
			}
			if role.StoreId != user.StoreId {
				return apperror.New(errcode.Forbidden, apperror.WithMsg("角色不属于当前企业"))
			}
			if role.IsSuper == 1 && user.IsSuper != 1 {
				return apperror.New(errcode.Forbidden, apperror.WithMsg("普通用户不能绑定超管角色"))
			}
		}

		filter := &RbacUserRoleFilterField{UserId: user.ID}
		existingRoles, err := s.userRoleRepo.FindAll(ctx, filter, nil, nil,
			baserepo.WithScopes(nil),
			baserepo.WithDB[*baserepo.QueryConfig](tx),
		)
		if err != nil {
			return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("设置用户角色失败"))
		}

		existingMap := make(map[uint32]bool)
		for _, role := range existingRoles {
			existingMap[role.RoleId] = true
		}

		newMap := make(map[uint32]bool)
		for _, id := range req.RoleIDs {
			newMap[id] = true
		}

		deleteIds := make([]uint32, 0)
		for _, role := range existingRoles {
			if !newMap[role.RoleId] {
				deleteIds = append(deleteIds, role.ID)
			}
		}
		if len(deleteIds) > 0 {
			if err := s.userRoleRepo.Delete(ctx, deleteIds, baserepo.WithDB[*baserepo.DeleteConfig](tx)); err != nil {
				return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("设置用户角色失败"))
			}
		}

		createUserRoles := make([]*RbacUserRole, 0)
		for _, roleID := range req.RoleIDs {
			if !existingMap[roleID] {
				createUserRoles = append(createUserRoles, &RbacUserRole{
					UserId:  user.ID,
					RoleId:  roleID,
					StoreId: req.StoreId,
				})
			}
		}
		if err := s.userRoleRepo.CreateBatch(ctx, createUserRoles, baserepo.WithDB[*baserepo.CreateConfig](tx)); err != nil {
			return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("设置用户角色失败"))
		}
		return nil
	})
}

// checkUsername 检查用户名
func (s *RbacUserService) checkUsername(ctx context.Context, username string, storeId uint32) error {
	filter := &RbacUserUsernameFilterField{Username: username, StoreId: storeId}
	_, err := s.repo.FindOne(ctx, filter)
	if err != nil {
		if errors.Is(err, baserepo.ErrRecordNotFound) {
			return nil
		}
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("检查用户名失败"))
	}
	return apperror.New(errcode.Conflict, apperror.WithMsg("用户名已存在"))
}

// ResetPassword 重置用户密码
func (s *RbacUserService) ResetPassword(ctx context.Context, req *RbacUserResetPasswordRequest) (string, error) {
	user, err := s.repo.FindOne(ctx, req.ID, baserepo.WithScopes(nil))
	if err != nil {
		if errors.Is(err, baserepo.ErrRecordNotFound) {
			return "", apperror.New(errcode.NotFound, apperror.WithMsg("用户不存在"))
		}
		return "", apperror.Wrap(errcode.Internal, err, apperror.WithMsg("重置密码失败"))
	}
	if user.StoreId != req.StoreId {
		return "", apperror.New(errcode.NotFound, apperror.WithMsg("用户不存在"))
	}

	newPassword := helper.RandomStringWithSymbols(12)
	hashedPassword, err := helper.HashPassword(newPassword)
	if err != nil {
		return "", apperror.Wrap(errcode.Internal, err, apperror.WithMsg("密码加密失败"))
	}

	updateData := map[string]any{"password": hashedPassword}
	if err := s.repo.Updates(ctx, user, updateData); err != nil {
		return "", apperror.Wrap(errcode.Internal, err, apperror.WithMsg("重置密码失败"))
	}
	return newPassword, nil
}

// RbacMenuService 菜单服务
type RbacMenuService struct {
	repo    *RbacMenuRepository
	apiRepo *RbacMenuApiRepository
	Db      *gorm.DB
}

// NewRbacMenuService 创建菜单服务
func NewRbacMenuService(repo *RbacMenuRepository, apiRepo *RbacMenuApiRepository, db *gorm.DB) *RbacMenuService {
	return &RbacMenuService{repo: repo, apiRepo: apiRepo, Db: db}
}

// FindTreeList 获取菜单树
func (s *RbacMenuService) FindTreeList(ctx context.Context) ([]*RbacMenu, error) {
	list, err := s.repo.FindAll(ctx, nil, nil, nil, baserepo.WithScopes(nil))
	if err != nil {
		return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("获取菜单列表失败"))
	}
	if len(list) > 0 {
		rbacMenuList := RbacMenuList{}
		for _, menu := range list {
			rbacMenuList = append(rbacMenuList, menu)
		}
		list = rbacMenuList.Tree()
	}
	return list, nil
}

// Create 创建菜单
func (s *RbacMenuService) Create(ctx context.Context, req *RbacMenuCreateRequest) error {
	if err := s.checkName(ctx, req.Name); err != nil {
		return err
	}
	item := &RbacMenu{
		Type:       req.Type,
		Name:       req.Name,
		ParentId:   req.ParentId,
		Path:       req.Path,
		Sort:       req.Sort,
		IsPage:     req.IsPage,
		ModuleKey:  req.ModuleKey,
		ActionMark: req.ActionMark,
	}
	if err := s.repo.Create(ctx, item); err != nil {
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("创建菜单失败"))
	}
	return nil
}

// Update 更新菜单
func (s *RbacMenuService) Update(ctx context.Context, req *RbacMenuUpdateRequest) error {
	item, err := s.repo.FindOne(ctx, req.ID, baserepo.WithScopes(nil))
	if err != nil {
		if errors.Is(err, baserepo.ErrRecordNotFound) {
			return apperror.New(errcode.NotFound)
		}
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("更新菜单失败"))
	}
	if item.Name != req.Name {
		if err := s.checkName(ctx, req.Name); err != nil {
			return err
		}
	}

	updateData := map[string]any{
		"type":        req.Type,
		"name":        req.Name,
		"path":        req.Path,
		"sort":        req.Sort,
		"module_key":  req.ModuleKey,
		"parent_id":   req.ParentId,
		"is_page":     req.IsPage,
		"action_mark": req.ActionMark,
	}
	if err := s.repo.Updates(ctx, item, updateData); err != nil {
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("更新菜单失败"))
	}
	return nil
}

// Delete 删除菜单
func (s *RbacMenuService) Delete(ctx context.Context, req *RbacMenuDeleteRequest) error {
	item, err := s.repo.FindOne(ctx, req.ID, baserepo.WithScopes(nil))
	if err != nil {
		if errors.Is(err, baserepo.ErrRecordNotFound) {
			return apperror.New(errcode.NotFound)
		}
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("删除菜单失败"))
	}
	if err := s.repo.Delete(ctx, item.ID); err != nil {
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("删除菜单失败"))
	}
	return nil
}

// Sync 同步菜单
func (s *RbacMenuService) Sync(ctx context.Context, req []RbacMenuSyncRequest) error {
	return s.Db.Transaction(func(tx *gorm.DB) error {
		list, err := s.repo.FindAll(ctx, &RbacMenuFilterField{Type: 10}, nil, nil,
			baserepo.WithDB[*baserepo.QueryConfig](tx),
			baserepo.WithScopes(nil),
		)
		if err != nil {
			return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("同步菜单失败"))
		}

		menuMap := map[string]*RbacMenu{}
		for _, item := range list {
			menuMap[item.Path] = item
		}

		if err := s.SyncMenuList(ctx, req, 0, menuMap, tx); err != nil {
			return err
		}
		for _, item := range menuMap {
			if err := s.repo.Delete(ctx, item.ID, baserepo.WithDB[*baserepo.DeleteConfig](tx)); err != nil {
				return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("同步菜单失败"))
			}
		}
		return nil
	})
}

// SyncMenuList 递归同步菜单
func (s *RbacMenuService) SyncMenuList(ctx context.Context, req []RbacMenuSyncRequest, parentId uint32, menuMap map[string]*RbacMenu, tx *gorm.DB) error {
	for _, item := range req {
		menu, exists := menuMap[item.Path]
		if exists {
			updateData := map[string]any{
				"module_key": item.ModuleKey,
				"sort":       item.Sort,
				"type":       item.Type,
				"name":       item.Name,
				"is_page":    item.IsPage,
				"parent_id":  parentId,
			}
			if err := s.repo.Updates(ctx, menu, updateData, baserepo.WithDB[*baserepo.UpdateConfig](tx)); err != nil {
				return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("同步菜单失败"))
			}
			delete(menuMap, item.Path)
		} else {
			menu = &RbacMenu{
				Type:      item.Type,
				Name:      item.Name,
				Path:      item.Path,
				IsPage:    item.IsPage,
				ModuleKey: item.ModuleKey,
				ParentId:  parentId,
				Sort:      item.Sort,
			}
			if err := s.repo.Create(ctx, menu, baserepo.WithDB[*baserepo.CreateConfig](tx)); err != nil {
				return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("同步菜单失败"))
			}
		}
		if len(item.Children) > 0 {
			if err := s.SyncMenuList(ctx, item.Children, menu.ID, menuMap, tx); err != nil {
				return err
			}
		}
	}
	return nil
}

// checkName 检查菜单名称
func (s *RbacMenuService) checkName(ctx context.Context, name string) error {
	filter := &RbacMenuFilterField{Name: name}
	_, err := s.repo.FindOne(ctx, filter, baserepo.WithScopes(nil))
	if err != nil {
		if errors.Is(err, baserepo.ErrRecordNotFound) {
			return nil
		}
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("检查菜单名称失败"))
	}
	return apperror.New(errcode.Conflict, apperror.WithMsg("菜单名已存在"))
}

// FindApiList 查询菜单关联的api
func (s *RbacMenuService) FindApiList(ctx context.Context, req *RbacMenuApiListRequest) (*MenuApiRelationResponse, error) {
	filter := &RbacMenuApiFilterField{MenuId: req.MenuID}
	list, err := s.apiRepo.FindAll(ctx, filter, nil, nil, baserepo.WithScopes(nil))
	if err != nil {
		return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("获取菜单API关联失败"))
	}
	apiIds := make([]uint32, 0)
	for _, item := range list {
		apiIds = append(apiIds, item.ApiId)
	}
	return &MenuApiRelationResponse{
		ApiIds: apiIds,
		MenuId: req.MenuID,
	}, nil
}

// SaveApiList 保存菜单权限
func (s *RbacMenuService) SaveApiList(ctx context.Context, req *RbacMenuApiSaveRequest) error {
	return s.Db.Transaction(func(tx *gorm.DB) error {
		filter := &RbacMenuApiFilterField{MenuId: req.MenuID}
		list, err := s.apiRepo.FindAll(ctx, filter, nil, nil,
			baserepo.WithDB[*baserepo.QueryConfig](tx),
			baserepo.WithScopes(nil),
		)
		if err != nil {
			return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("保存菜单API关联失败"))
		}

		apiIds := make(map[uint32]bool)
		for _, item := range list {
			apiIds[item.ApiId] = true
		}

		for _, apiId := range req.ApiIds {
			if !apiIds[apiId] {
				item := &RbacMenuApi{MenuId: req.MenuID, ApiId: apiId}
				err := s.apiRepo.Create(ctx, item, baserepo.WithDB[*baserepo.CreateConfig](tx))
				if err != nil {
					return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("保存菜单API关联失败"))
				}
			} else {
				delete(apiIds, apiId)
			}
		}

		for apiId := range apiIds {
			filter := &RbacMenuApiFilterField{ApiId: apiId, MenuId: req.MenuID}
			err := s.apiRepo.Delete(ctx, filter, baserepo.WithDB[*baserepo.DeleteConfig](tx))
			if err != nil {
				return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("保存菜单API关联失败"))
			}
		}
		return nil
	})
}

// RbacRoleService 角色服务
type RbacRoleService struct {
	db           *gorm.DB
	repo         *RbacRoleRepository
	roleMenuRepo *RbacRoleMenuRepository
}

// NewRbacRoleService 创建角色服务
func NewRbacRoleService(repo *RbacRoleRepository, roleMenuRepo *RbacRoleMenuRepository, db *gorm.DB) *RbacRoleService {
	return &RbacRoleService{db: db, repo: repo, roleMenuRepo: roleMenuRepo}
}

// FindTreeList 获取角色树形列表
func (s *RbacRoleService) FindTreeList(ctx context.Context, req *RbacRoleListRequest) ([]*RbacRole, error) {
	filter := &RbacRoleFilterField{
		StoreId:  req.StoreId,
		RoleName: req.RoleName,
		IsSuper:  req.IsSuper,
	}
	list, err := s.repo.FindAll(ctx, filter, nil, nil,
		baserepo.WithScopes(nil),
		baserepo.WithPreloads("RbacRoleMenu.RbacMenu"),
	)
	if err != nil {
		return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("获取角色列表失败"))
	}
	if len(list) > 0 {
		rbacRoleList := RbacRoleList{}
		for _, role := range list {
			rbacRoleList = append(rbacRoleList, role)
		}
		list = rbacRoleList.Tree()
	}
	return list, nil
}

// FindList 获取角色分页列表(平台端跨企业视图使用，不构建树形结构)
func (s *RbacRoleService) FindList(ctx context.Context, req *RbacRolePageListRequest) (*ListResult, error) {
	result := &ListResult{List: []*RbacRole{}, Total: 0}

	filter := &RbacRoleFilterField{
		StoreId:  req.StoreId,
		RoleName: req.RoleName,
		IsSuper:  req.IsSuper,
	}

	pagination := baserepo.NewPagination(req.Page, req.Limit)
	orders := baserepo.Orders{
		{Field: "sort", Sort: "asc"},
		{Field: "id", Sort: "desc"},
	}

	total, err := s.repo.Count(ctx, filter, baserepo.WithScopes(nil))
	if err != nil {
		return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("获取角色列表失败"))
	}
	if total == 0 {
		return result, nil
	}
	result.Total = total

	list, err := s.repo.FindAll(ctx, filter, pagination, orders,
		baserepo.WithScopes(nil),
		baserepo.WithPreloads("RbacRoleMenu.RbacMenu"),
	)
	if err != nil {
		return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("获取角色列表失败"))
	}
	result.List = list
	return result, nil
}

// Create 创建角色
func (s *RbacRoleService) Create(ctx context.Context, req *RbacRoleCreateRequest) error {
	if err := s.checkName(ctx, req.RoleName, req.StoreId); err != nil {
		return err
	}
	if req.IsSuper == 1 {
		if req.StoreId == 0 {
			return apperror.New(errcode.InvalidInput, apperror.WithMsg("创建超管角色必须指定所属企业"))
		}
		if err := s.checkSuperUnique(ctx, req.StoreId, 0); err != nil {
			return err
		}
	}
	if req.ParentId > 0 {
		if err := s.checkParent(ctx, req.ParentId, req.StoreId); err != nil {
			return err
		}
	}
	item := &RbacRole{
		RoleName: req.RoleName,
		ParentId: req.ParentId,
		Sort:     req.Sort,
		StoreId:  req.StoreId,
		IsSuper:  req.IsSuper,
	}
	if err := s.repo.Create(ctx, item); err != nil {
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("创建角色失败"))
	}
	return nil
}

// Update 更新角色
func (s *RbacRoleService) Update(ctx context.Context, req *RbacRoleUpdateRequest) error {
	item, err := s.repo.FindOne(ctx, req.ID, baserepo.WithScopes(nil))
	if err != nil {
		if errors.Is(err, baserepo.ErrRecordNotFound) {
			return apperror.New(errcode.NotFound)
		}
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("更新角色失败"))
	}
	if item.StoreId != req.StoreId {
		return apperror.New(errcode.NotFound)
	}
	if item.RoleName != req.RoleName {
		if err := s.checkName(ctx, req.RoleName, req.StoreId); err != nil {
			return err
		}
	}
	if req.IsSuper == 1 {
		if err := s.checkSuperUnique(ctx, req.StoreId, req.ID); err != nil {
			return err
		}
	}

	updateData := map[string]any{
		"role_name": req.RoleName,
		"parent_id": req.ParentId,
		"sort":      req.Sort,
		"store_id":  req.StoreId,
		"is_super":  req.IsSuper,
	}
	if err := s.repo.Updates(ctx, item, updateData); err != nil {
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("更新角色失败"))
	}
	return nil
}

// Delete 删除角色
func (s *RbacRoleService) Delete(ctx context.Context, req *RbacRoleDeleteRequest) error {
	item, err := s.repo.FindOne(ctx, req.ID, baserepo.WithScopes(nil))
	if err != nil {
		if errors.Is(err, baserepo.ErrRecordNotFound) {
			return apperror.New(errcode.NotFound)
		}
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("删除角色失败"))
	}
	if item.StoreId != req.StoreId {
		return apperror.New(errcode.NotFound)
	}
	if err := s.repo.Delete(ctx, item.ID); err != nil {
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("删除角色失败"))
	}
	return nil
}

// SetMenus 设置角色菜单
func (s *RbacRoleService) SetMenus(ctx context.Context, req *RbacRoleMenuSetRequest) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		role, err := s.repo.FindOne(ctx, req.RoleID, baserepo.WithDB[*baserepo.QueryConfig](tx), baserepo.WithScopes(nil))
		if err != nil {
			if errors.Is(err, baserepo.ErrRecordNotFound) {
				return apperror.New(errcode.NotFound, apperror.WithMsg("角色不存在"))
			}
			return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("设置角色菜单失败"))
		}
		if role.StoreId != req.StoreId {
			return apperror.New(errcode.NotFound, apperror.WithMsg("角色不存在"))
		}

		filter := &RbacRoleMenuFilterField{RoleId: role.ID}
		existingMenus, err := s.roleMenuRepo.FindAll(ctx, filter, nil, nil,
			baserepo.WithDB[*baserepo.QueryConfig](tx),
			baserepo.WithScopes(nil),
		)
		if err != nil {
			return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("设置角色菜单失败"))
		}

		existingMap := make(map[uint32]bool)
		for _, menu := range existingMenus {
			existingMap[menu.MenuId] = true
		}

		newMap := make(map[uint32]bool)
		for _, id := range req.MenuIDs {
			newMap[id] = true
		}

		deleteIds := make([]uint32, 0)
		for _, menu := range existingMenus {
			if !newMap[menu.MenuId] {
				deleteIds = append(deleteIds, menu.ID)
			}
		}
		if len(deleteIds) > 0 {
			if err := s.roleMenuRepo.Delete(ctx, deleteIds, baserepo.WithDB[*baserepo.DeleteConfig](tx)); err != nil {
				return err
			}
		}

		roleMenus := make([]*RbacRoleMenu, 0)
		for _, menuID := range req.MenuIDs {
			if !existingMap[menuID] {
				roleMenus = append(roleMenus, &RbacRoleMenu{
					RoleId:  role.ID,
					MenuId:  menuID,
					StoreId: req.StoreId,
				})
			}
		}
		if err := s.roleMenuRepo.CreateBatch(ctx, roleMenus, baserepo.WithDB[*baserepo.CreateConfig](tx)); err != nil {
			return err
		}
		return nil
	})
}

// checkName 检查角色名称
func (s *RbacRoleService) checkName(ctx context.Context, name string, StoreId uint32) error {
	_, err := s.repo.FindByName(ctx, name, StoreId)
	if err != nil {
		if errors.Is(err, baserepo.ErrRecordNotFound) {
			return nil
		}
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("检查角色名称失败"))
	}
	return apperror.New(errcode.Conflict, apperror.WithMsg("角色名已存在"))
}

// checkSuperUnique 校验同一企业下超管角色唯一
func (s *RbacRoleService) checkSuperUnique(ctx context.Context, storeId uint32, excludeID uint32) error {
	superRole, err := s.repo.FindSuper(ctx, storeId)
	if err != nil {
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("检查超管角色失败"))
	}
	if superRole.ID != 0 && superRole.ID != excludeID {
		return apperror.New(errcode.Conflict, apperror.WithMsg("该企业已存在超管角色"))
	}
	return nil
}

// checkParent 检查父级角色
func (s *RbacRoleService) checkParent(ctx context.Context, parentId uint32, StoreId uint32) error {
	parent, err := s.repo.FindOne(ctx, parentId)
	if err != nil {
		if errors.Is(err, baserepo.ErrRecordNotFound) {
			return apperror.New(errcode.NotFound, apperror.WithMsg("父级角色不存在"))
		}
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("检查父级角色失败"))
	}
	if parent.StoreId != StoreId {
		return apperror.New(errcode.Forbidden, apperror.WithMsg("父级角色不属于当前企业"))
	}
	return nil
}

// RbacApiService 接口服务
type RbacApiService struct {
	repo *RbacApiRepository
}

// NewRbacApiService 创建接口服务
func NewRbacApiService(repo *RbacApiRepository) *RbacApiService {
	return &RbacApiService{repo: repo}
}

// FindTreeList 获取接口树列表
func (s *RbacApiService) FindTreeList(ctx context.Context) ([]*RbacApi, error) {
	list, err := s.repo.FindAll(ctx, nil, nil, nil, baserepo.WithScopes(nil))
	if err != nil {
		return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("获取接口列表失败"))
	}
	if len(list) > 0 {
		rbacApiList := RbacApiList{}
		for _, api := range list {
			rbacApiList = append(rbacApiList, api)
		}
		list = rbacApiList.Tree()
	}
	return list, nil
}

// Create 创建接口
func (s *RbacApiService) Create(ctx context.Context, req *RbacApiCreateRequest) error {
	if err := s.checkName(ctx, req.Name); err != nil {
		return err
	}
	item := &RbacApi{
		Name:     req.Name,
		ParentId: req.ParentId,
		Url:      req.Url,
		Sort:     req.Sort,
	}
	if err := s.repo.Create(ctx, item); err != nil {
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("创建接口失败"))
	}
	return nil
}

// Update 更新接口
func (s *RbacApiService) Update(ctx context.Context, req *RbacApiUpdateRequest) error {
	item, err := s.repo.FindOne(ctx, req.ID, baserepo.WithScopes(nil))
	if err != nil {
		if errors.Is(err, baserepo.ErrRecordNotFound) {
			return apperror.New(errcode.NotFound)
		}
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("更新接口失败"))
	}
	if item.Name != req.Name {
		if err := s.checkName(ctx, req.Name); err != nil {
			return err
		}
	}

	updateData := map[string]any{
		"name":      req.Name,
		"parent_id": req.ParentId,
		"url":       req.Url,
		"sort":      req.Sort,
	}
	if err := s.repo.Updates(ctx, item, updateData); err != nil {
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("更新接口失败"))
	}
	return nil
}

// Delete 删除接口
func (s *RbacApiService) Delete(ctx context.Context, req *RbacApiDeleteRequest) error {
	item, err := s.repo.FindOne(ctx, req.ID, baserepo.WithScopes(nil))
	if err != nil {
		if errors.Is(err, baserepo.ErrRecordNotFound) {
			return apperror.New(errcode.NotFound)
		}
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("删除接口失败"))
	}
	if err := s.repo.Delete(ctx, item.ID); err != nil {
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("删除接口失败"))
	}
	return nil
}

// checkName 检查接口名称
func (s *RbacApiService) checkName(ctx context.Context, name string) error {
	filter := &RbacApiFilterField{Name: name}
	_, err := s.repo.FindOne(ctx, filter, baserepo.WithScopes(nil))
	if err != nil {
		if errors.Is(err, baserepo.ErrRecordNotFound) {
			return nil
		}
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("检查接口名称失败"))
	}
	return apperror.New(errcode.Conflict, apperror.WithMsg("接口名已存在"))
}

// RbacStoreService 企业服务
type RbacStoreService struct {
	repo *RbacStoreRepository
}

// NewRbacStoreService 创建企业服务
func NewRbacStoreService(repo *RbacStoreRepository) *RbacStoreService {
	return &RbacStoreService{repo: repo}
}

// FindList 获取企业列表
func (s *RbacStoreService) FindList(ctx context.Context, req *RbacStoreListRequest) (*ListResult, error) {
	result := &ListResult{List: []*RbacStore{}, Total: 0}

	filter := &RbacStoreFilterField{
		Name:      req.Name,
		IsRecycle: req.IsRecycle,
	}

	pagination := baserepo.NewPagination(req.Page, req.Limit)
	orders := baserepo.Orders{
		{Field: "sort", Sort: "asc"},
		{Field: "id", Sort: "desc"},
	}

	total, err := s.repo.Count(ctx, filter, baserepo.WithScopes(nil))
	if err != nil {
		return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("获取企业列表失败"))
	}
	if total == 0 {
		return result, nil
	}
	result.Total = total

	list, err := s.repo.FindAll(ctx, filter, pagination, orders, baserepo.WithScopes(nil))
	if err != nil {
		return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("获取企业列表失败"))
	}
	result.List = list
	return result, nil
}

// Create 创建企业
func (s *RbacStoreService) Create(ctx context.Context, req *RbacStoreCreateRequest) error {
	if err := s.checkName(ctx, req.Name); err != nil {
		return err
	}
	item := &RbacStore{
		Name:         req.Name,
		ShortName:    req.ShortName,
		Contact:      req.Contact,
		ContactPhone: req.ContactPhone,
		Description:  req.Description,
		LogoImageID:  req.LogoImageID,
		Sort:         req.Sort,
	}
	if err := s.repo.Create(ctx, item); err != nil {
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("创建企业失败"))
	}
	return nil
}

// checkName 检查企业名称
func (s *RbacStoreService) checkName(ctx context.Context, name string) error {
	filter := &RbacStoreNameFilterField{Name: name}
	_, err := s.repo.FindOne(ctx, filter, baserepo.WithScopes(nil))
	if err != nil {
		if errors.Is(err, baserepo.ErrRecordNotFound) {
			return nil
		}
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("检查企业名称失败"))
	}
	return apperror.New(errcode.Conflict, apperror.WithMsg("企业名已存在"))
}

// Update 更新企业信息
func (s *RbacStoreService) Update(ctx context.Context, req *RbacStoreUpdateRequest) error {
	item, err := s.repo.FindOne(ctx, req.ID, baserepo.WithScopes(nil))
	if err != nil {
		if errors.Is(err, baserepo.ErrRecordNotFound) {
			return apperror.New(errcode.NotFound)
		}
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("更新企业失败"))
	}
	if item.Name != req.Name {
		if err := s.checkName(ctx, req.Name); err != nil {
			return err
		}
	}

	updateData := map[string]any{
		"contact":       req.Contact,
		"contact_phone": req.ContactPhone,
		"description":   req.Description,
		"name":          req.Name,
		"short_name":    req.ShortName,
		"sort":          req.Sort,
		"logo_image_id": req.LogoImageID,
	}
	if err := s.repo.Updates(ctx, item, updateData); err != nil {
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("更新企业失败"))
	}
	return nil
}

// Delete 删除企业信息
func (s *RbacStoreService) Delete(ctx context.Context, req *RbacStoreDeleteRequest) error {
	item, err := s.repo.FindOne(ctx, req.ID, baserepo.WithScopes(nil))
	if err != nil {
		if errors.Is(err, baserepo.ErrRecordNotFound) {
			return apperror.New(errcode.NotFound)
		}
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("删除企业失败"))
	}
	if err := s.repo.Delete(ctx, item.ID); err != nil {
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("删除企业失败"))
	}
	return nil
}

// Recycle 将企业移入回收站
func (s *RbacStoreService) Recycle(ctx context.Context, req *RbacStoreDeleteRequest) error {
	item, err := s.repo.FindOne(ctx, req.ID)
	if err != nil {
		if errors.Is(err, baserepo.ErrRecordNotFound) {
			return apperror.New(errcode.NotFound)
		}
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("移入回收站失败"))
	}

	updateData := map[string]any{"is_recycle": 1}
	if err := s.repo.Updates(ctx, item, updateData); err != nil {
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("移入回收站失败"))
	}
	return nil
}

// Restore 将企业从回收站恢复
func (s *RbacStoreService) Restore(ctx context.Context, req *RbacStoreDeleteRequest) error {
	item, err := s.repo.FindOne(ctx, req.ID)
	if err != nil {
		if errors.Is(err, baserepo.ErrRecordNotFound) {
			return apperror.New(errcode.NotFound)
		}
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("从回收站恢复失败"))
	}

	updateData := map[string]any{"is_recycle": 0}
	if err := s.repo.Updates(ctx, item, updateData); err != nil {
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("从回收站恢复失败"))
	}
	return nil
}
