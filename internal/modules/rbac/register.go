package rbac

import (
	"github.com/241x/zero-kit/bind"
	"github.com/241x/zero-web/middleware"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// buildAll 构建 rbac 模块所有服务。
func buildAll(db *gorm.DB, binder *bind.Binder, config Config, rdb *redis.Client, captcha CaptchaVerifier) (*handler, *AuthMiddleware) {
	userRepo := NewRbacUserRepository(db)
	userRoleRepo := NewRbacUserRoleRepository(db)
	menuRepo := NewRbacMenuRepository(db)
	menuApiRepo := NewRbacMenuApiRepository(db)
	apiRepo := NewRbacApiRepository(db)
	roleRepo := NewRbacRoleRepository(db)
	roleMenuRepo := NewRbacRoleMenuRepository(db)
	storeRepo := NewRbacStoreRepository(db)

	authServ := NewAuthService(
		userRepo, apiRepo, roleRepo, menuRepo,
		roleMenuRepo, userRoleRepo, menuApiRepo,
		config, rdb, captcha,
	)

	authMid := NewAuthMiddleware(config, authServ)
	menuServ := NewRbacMenuService(menuRepo, menuApiRepo, db)
	apiServ := NewRbacApiService(apiRepo)
	roleServ := NewRbacRoleService(roleRepo, roleMenuRepo, db)
	userServ := NewRbacUserService(db, userRepo, userRoleRepo)
	storeServ := NewRbacStoreService(storeRepo)

	h := newHandler(binder, authServ, config, menuServ, apiServ, roleServ, userServ, storeServ, authMid)

	return h, authMid
}

// RegisterAdmin 注册 rbac 模块路由。public 注册公开路由，protected 注册需认证的路由并挂载 JWT 中间件。
func RegisterAdmin(public, r *gin.RouterGroup, db *gorm.DB, binder *bind.Binder, config Config, rdb *redis.Client, captcha CaptchaVerifier) {
	h, authMid := buildAll(db, binder, config, rdb, captcha)

	public.POST("/login", h.login)
	public.POST("/refresh-token", h.refreshToken)

	r.Use(middleware.JWTGuard(config.HmacSecret))
	r.Use(authMid.LoadUser())
	r.Use(authMid.CheckAPIPermission())

	r.POST("/logout", h.logout)
	r.POST("/change-password", h.changePassword)
	r.POST("/permissions", h.permissions)

	r.POST("/rbac/role/list", h.roleList)
	r.POST("/rbac/role/create", h.roleCreate)
	r.POST("/rbac/role/update", h.roleUpdate)
	r.POST("/rbac/role/delete", h.roleDelete)
	r.POST("/rbac/role/set-menus", h.roleSetMenus)

	r.POST("/rbac/user/list", h.userList)
	r.POST("/rbac/user/create", h.userCreate)
	r.POST("/rbac/user/update", h.userUpdate)
	r.POST("/rbac/user/delete", h.userDelete)
	r.POST("/rbac/user/set-roles", h.userSetRoles)
	r.POST("/rbac/user/reset-password", h.userResetPassword)
}

// RegisterPlatform 注册平台端路由。不挂载 RBAC 权限中间件，由平台端的 RequireRole 中间件控制访问权限。
func RegisterPlatform(r *gin.RouterGroup, db *gorm.DB, binder *bind.Binder, config Config, rdb *redis.Client) {
	h, _ := buildAll(db, binder, config, rdb, nil)

	r.POST("/rbac/menu/list", h.menuList)
	r.POST("/rbac/menu/create", h.menuCreate)
	r.POST("/rbac/menu/update", h.menuUpdate)
	r.POST("/rbac/menu/delete", h.menuDelete)
	r.POST("/rbac/menu/sync", h.menuSync)
	r.POST("/rbac/menu/api/list", h.menuApiList)
	r.POST("/rbac/menu/api/save", h.menuApiSave)

	r.POST("/rbac/api/list", h.apiList)
	r.POST("/rbac/api/create", h.apiCreate)
	r.POST("/rbac/api/update", h.apiUpdate)
	r.POST("/rbac/api/delete", h.apiDelete)

	r.POST("/rbac/store/list", h.storeList)
	r.POST("/rbac/store/create", h.storeCreate)
	r.POST("/rbac/store/update", h.storeUpdate)
	r.POST("/rbac/store/delete", h.storeDelete)
	r.POST("/rbac/store/recycle", h.storeRecycle)
	r.POST("/rbac/store/restore", h.storeRestore)

	r.POST("/rbac/role/list", h.platformRoleList)
	r.POST("/rbac/role/create", h.roleCreate)
	r.POST("/rbac/role/update", h.roleUpdate)
	r.POST("/rbac/role/delete", h.roleDelete)
	r.POST("/rbac/role/set-menus", h.roleSetMenus)

	r.POST("/rbac/user/list", h.userList)
	r.POST("/rbac/user/create", h.userCreate)
	r.POST("/rbac/user/update", h.userUpdate)
	r.POST("/rbac/user/delete", h.userDelete)
	r.POST("/rbac/user/reset-password", h.userResetPassword)
}
