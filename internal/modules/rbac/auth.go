package rbac

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/241x/zero-kit/apperror"
	"github.com/241x/zero-kit/baserepo"
	"github.com/241x/zero-kit/helper"
	"github.com/241x/zero-web/ctxkeys"
	"github.com/241x/zero-web/errcode"
	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

const (
	redisAdminLoginKey        = "ZAG:ADMIN:LOGIN"
	redisAdminRefreshTokenKey = "ZAG:ADMIN:REFRESH:TOKEN"
)

// CaptchaVerifier 验证码验证接口（由宿主注入）
type CaptchaVerifier interface {
	Verify(ctx context.Context, captchaID, captchaCode string) error
}

// AuthService 管理员认证服务
type AuthService struct {
	userRepo     *RbacUserRepository
	apiRepo      *RbacApiRepository
	roleRepo     *RbacRoleRepository
	menuRepo     *RbacMenuRepository
	roleMenuRepo *RbacRoleMenuRepository
	userRoleRepo *RbacUserRoleRepository
	menuApiRepo  *RbacMenuApiRepository
	cfg          Config
	rdb          *redis.Client
	captcha      CaptchaVerifier
}

// NewAuthService 创建认证服务
func NewAuthService(
	userRepo *RbacUserRepository,
	apiRepo *RbacApiRepository,
	roleRepo *RbacRoleRepository,
	menuRepo *RbacMenuRepository,
	roleMenuRepo *RbacRoleMenuRepository,
	userRoleRepo *RbacUserRoleRepository,
	menuApiRepo *RbacMenuApiRepository,
	cfg Config,
	rdb *redis.Client,
	captcha CaptchaVerifier,
) *AuthService {
	return &AuthService{
		userRepo:     userRepo,
		apiRepo:      apiRepo,
		roleRepo:     roleRepo,
		menuRepo:     menuRepo,
		roleMenuRepo: roleMenuRepo,
		userRoleRepo: userRoleRepo,
		menuApiRepo:  menuApiRepo,
		cfg:          cfg,
		rdb:          rdb,
		captcha:      captcha,
	}
}

// Login 系统登录
func (s *AuthService) Login(ctx context.Context, req *AuthLoginRequest) (*AdminLoginResponse, string, error) {
	if err := s.captcha.Verify(ctx, req.CaptchaID, req.CaptchaCode); err != nil {
		return nil, "", err
	}

	filter := &RbacUserUsernameFilterField{Username: req.Username}
	item, err := s.userRepo.FindOne(ctx, filter)
	if err != nil {
		if errors.Is(err, baserepo.ErrRecordNotFound) {
			return nil, "", apperror.New(errcode.InvalidInput, apperror.WithMsg("用户名或密码错误"))
		}
		return nil, "", apperror.Wrap(errcode.Internal, err, apperror.WithMsg("登录失败"))
	}

	if item.Password == "" {
		return nil, "", apperror.New(errcode.InvalidInput, apperror.WithMsg("该账号尚未设置密码，请联系平台重置密码"))
	}

	ok, err := helper.CheckPassword(req.Password, item.Password)
	if err != nil {
		return nil, "", apperror.Wrap(errcode.Internal, err, apperror.WithMsg("验证密码失败"))
	}
	if !ok {
		return nil, "", apperror.New(errcode.InvalidInput, apperror.WithMsg("用户名或密码错误"))
	}

	refreshToken, err := s.getRefreshToken()
	if err != nil {
		return nil, "", err
	}

	itemBytes, err := json.Marshal(item)
	if err != nil {
		return nil, "", apperror.Wrap(errcode.Internal, err, apperror.WithMsg("登录失败"))
	}

	result := s.rdb.Set(ctx,
		fmt.Sprintf("%s:%s", redisAdminRefreshTokenKey, refreshToken),
		itemBytes,
		time.Duration(s.cfg.RefreshTokenTtl)*time.Second)
	if result.Err() != nil {
		return nil, "", apperror.Wrap(errcode.Internal, result.Err(), apperror.WithMsg("登录失败"))
	}

	tokenString, err := s.getAccessToken(item)
	if err != nil {
		return nil, "", apperror.Wrap(errcode.Internal, err, apperror.WithMsg("登录失败"))
	}

	return &AdminLoginResponse{
		Token: tokenString,
		Ttl:   s.cfg.RefreshTokenTtl,
		User:  item,
	}, refreshToken, nil
}

// RefreshToken 刷新Token
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*AdminLoginResponse, error) {
	itemBytes, err := s.rdb.Get(ctx,
		fmt.Sprintf("%s:%s", redisAdminRefreshTokenKey, refreshToken)).Bytes()
	if err != nil {
		return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("刷新令牌失败"))
	}

	item := &RbacUser{}
	if err := json.Unmarshal(itemBytes, item); err != nil {
		return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("刷新令牌失败"))
	}

	token, err := s.getAccessToken(item)
	if err != nil {
		return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("刷新令牌失败"))
	}

	return &AdminLoginResponse{
		Token: token,
		Ttl:   s.cfg.AccessTokenTtl,
		User:  nil,
	}, nil
}

// getRefreshToken 获取刷新Token
func (s *AuthService) getRefreshToken() (string, error) {
	token := helper.StringMd5(fmt.Sprintf("%d", time.Now().UnixNano()) + helper.RandomString(16))
	return token, nil
}

// getAccessToken 获取访问Token
func (s *AuthService) getAccessToken(item *RbacUser) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iat":      time.Now().Unix(),
		"exp":      time.Now().Add(time.Duration(s.cfg.AccessTokenTtl) * time.Second).Unix(),
		"user_id":  item.ID,
		"store_id": item.StoreId,
	})

	tokenString, err := token.SignedString([]byte(s.cfg.HmacSecret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// GetUserInfo 获取用户信息
func (s *AuthService) GetUserInfo(ctx context.Context, userId uint32) (*RbacUser, error) {
	cacheKey := fmt.Sprintf("%s:%d", redisAdminLoginKey, userId)
	result := s.rdb.Get(ctx, cacheKey)

	if result.Err() == nil {
		itemBytes, err := result.Bytes()
		if err == nil {
			item := &RbacUser{}
			if err := json.Unmarshal(itemBytes, item); err == nil {
				return item, nil
			}
		}
	}

	user, err := s.userRepo.FindOne(ctx, userId)
	if err != nil {
		if errors.Is(err, baserepo.ErrRecordNotFound) {
			return nil, apperror.New(errcode.NotFound, apperror.WithMsg("用户不存在"))
		}
		return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("获取用户信息失败"))
	}

	userBytes, err := json.Marshal(user)
	if err != nil {
		return user, nil
	}
	s.rdb.Set(ctx, cacheKey, userBytes, time.Hour)

	return user, nil
}

// GetPermissions 获取用户菜单权限
func (s *AuthService) GetPermissions(ctx context.Context, req *AuthGetPermissionsRequest) ([]*RbacMenu, error) {
	user := ctxkeys.User(ctx).(*RbacUser)
	if user == nil {
		return nil, nil
	}

	roles, err := s.GetUserRoles(ctx, user.ID)
	if err != nil {
		return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("获取菜单权限失败"))
	}
	if len(roles) == 0 {
		return nil, nil
	}

	var menus []*RbacMenu
	for _, role := range roles {
		roleMenus, err := s.GetRoleMenus(ctx, role.ID)
		if err != nil {
			return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("获取菜单权限失败"))
		}
		menus = append(menus, roleMenus...)
	}

	uniqueMenus := make(map[uint32]*RbacMenu)
	for _, menu := range menus {
		uniqueMenus[menu.ID] = menu
	}

	var result []*RbacMenu
	for _, menu := range uniqueMenus {
		result = append(result, menu)
	}

	if req.IsTree {
		return RbacMenuList(result).Tree(), nil
	}
	return result, nil
}

// GetUserRoles 根据用户ID获取角色列表
func (s *AuthService) GetUserRoles(ctx context.Context, userID uint32) ([]*RbacRole, error) {
	filter := &RbacUserRoleFilterField{UserId: userID}
	userRoles, err := s.userRoleRepo.FindAll(ctx, filter, nil, nil)
	if err != nil {
		return nil, err
	}
	if len(userRoles) == 0 {
		return []*RbacRole{}, nil
	}

	var roleIDs []uint32
	for _, ur := range userRoles {
		roleIDs = append(roleIDs, ur.RoleId)
	}

	roles, err := s.roleRepo.FindAll(ctx, &RbacRoleFilterField{IDs: roleIDs}, nil, nil)
	if err != nil {
		return nil, err
	}
	return roles, nil
}

// GetRoleMenus 根据角色ID获取角色菜单
func (s *AuthService) GetRoleMenus(ctx context.Context, roleId uint32) ([]*RbacMenu, error) {
	roleMenus, err := s.roleMenuRepo.FindAll(ctx, &RbacRoleMenuFilterField{RoleId: roleId}, nil, nil)
	if err != nil {
		return nil, err
	}
	if len(roleMenus) == 0 {
		return []*RbacMenu{}, nil
	}

	var menuIDs []uint32
	for _, rm := range roleMenus {
		menuIDs = append(menuIDs, rm.MenuId)
	}

	menus, err := s.menuRepo.FindAll(ctx, &RbacMenuFilterField{IDs: menuIDs}, nil, nil, baserepo.WithScopes(nil))
	if err != nil {
		return nil, err
	}
	return menus, nil
}

// ChangePassword 修改密码
func (s *AuthService) ChangePassword(ctx context.Context, req *ChangePasswordRequest) error {
	user, _ := ctxkeys.User(ctx).(*RbacUser)
	if user == nil || user.ID == 0 {
		return apperror.New(errcode.NotFound, apperror.WithMsg("用户不存在"))
	}

	user, err := s.userRepo.FindOne(ctx, user.ID)
	if err != nil {
		if errors.Is(err, baserepo.ErrRecordNotFound) {
			return apperror.New(errcode.NotFound, apperror.WithMsg("用户不存在"))
		}
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("修改密码失败"))
	}

	ok, err := helper.CheckPassword(req.OldPassword, user.Password)
	if err != nil {
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("验证密码失败"))
	}
	if !ok {
		return apperror.New(errcode.InvalidInput, apperror.WithMsg("旧密码不正确"))
	}

	hashedPassword, err := helper.HashPassword(req.NewPassword)
	if err != nil {
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("修改密码失败"))
	}

	updateData := map[string]any{"password": hashedPassword}
	if err := s.userRepo.Updates(ctx, user, updateData); err != nil {
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("修改密码失败"))
	}
	return nil
}

// CheckAPIPermission 检查API权限
func (s *AuthService) CheckAPIPermission(ctx context.Context, user *RbacUser, apiPath string) (bool, error) {
	roles, err := s.GetUserRoles(ctx, user.ID)
	if err != nil {
		return false, err
	}
	if len(roles) == 0 {
		return false, nil
	}

	api, err := s.apiRepo.GetAPIByPath(ctx, apiPath)
	if err != nil {
		return false, err
	}
	if api == nil || api.ID == 0 {
		return false, nil
	}

	for _, role := range roles {
		hasPerm, err := s.CheckRoleAPIPermission(ctx, role.ID, api.ID)
		if err != nil {
			return false, err
		}
		if hasPerm {
			return true, nil
		}
	}
	return false, nil
}

// CheckRoleAPIPermission 检查角色是否有API权限
func (s *AuthService) CheckRoleAPIPermission(ctx context.Context, roleID uint32, apiID uint32) (bool, error) {
	roleMenus, err := s.roleMenuRepo.FindAll(ctx, &RbacRoleMenuFilterField{RoleId: roleID}, nil, nil)
	if err != nil {
		return false, err
	}
	if len(roleMenus) == 0 {
		return false, nil
	}

	var menuIDs []uint32
	for _, rm := range roleMenus {
		menuIDs = append(menuIDs, rm.MenuId)
	}

	filter := &RbacMenuApiFilterField{MenuIDs: menuIDs, ApiId: apiID}
	menuApis, err := s.menuApiRepo.FindAll(ctx, filter, nil, nil, baserepo.WithScopes(nil))
	if err != nil {
		return false, err
	}
	return len(menuApis) > 0, nil
}
