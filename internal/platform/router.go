package platform

import (
	"zero-backend/internal/modules/captcha"
	"zero-backend/internal/modules/platform/user"
	"zero-backend/internal/modules/rbac"
	"zero-backend/internal/modules/setting"
	"zero-backend/internal/modules/upload"
	"zero-backend/internal/provider"

	"github.com/241x/zero-kit/bind"
	"github.com/241x/zero-kit/logger"
	"github.com/241x/zero-web/middleware"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// NewGin 创建一个 gin.Engine 实例
func NewGin(
	log logger.Logger,
	db *gorm.DB,
	binder *bind.Binder,
	rdb *redis.Client,
) *gin.Engine {
	r := gin.Default()
	r.Use(middleware.CORS(provider.LoadPlatformCorsConfig()))
	r.Use(middleware.Trace(log))
	r.Use(middleware.RequestLog())

	public := r.Group("/api")
	protected := public.Group("")

	captchaSvc := provider.MustNewCaptchaService(rdb, provider.LoadPlatformCaptchaConfig())
	captcha.Register(public, binder, captchaSvc)

	authCfg := user.MustLoadConfig()
	authMid := user.Register(public, protected, db, binder, authCfg, rdb, captchaSvc)

	protected.Use(authMid.RequireRole(user.RoleSuperAdmin, user.RoleOperator))

	cfg := rbac.MustLoadPlatformConfig()
	rbac.RegisterPlatform(protected, db, binder, cfg, rdb)
	setting.RegisterPlatform(protected, db, binder)

	settingSvc := provider.NewSettingService(db)
	upload.RegisterPlatform(protected, db, binder, settingSvc)

	return r
}
