package admin

import (
	"net/http"

	"zero-backend/internal/modules/article"
	"zero-backend/internal/modules/captcha"
	"zero-backend/internal/modules/dashboard"
	"zero-backend/internal/modules/health"
	"zero-backend/internal/modules/rbac"
	"zero-backend/internal/modules/region"
	"zero-backend/internal/modules/setting"
	"zero-backend/internal/modules/upload"
	"zero-backend/internal/modules/user"
	"zero-backend/internal/modules/wecom/config"
	"zero-backend/internal/modules/wecom/contact"
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
	r.Use(middleware.CORS(provider.LoadAdminCorsConfig()))
	r.Use(middleware.Trace(log))
	r.Use(middleware.RequestLog())

	public := r.Group("/api")
	protected := public.Group("")

	captchaSvc := provider.MustNewCaptchaService(rdb, provider.LoadAdminCaptchaConfig())
	captcha.Register(public, binder, captchaSvc)

	authCfg := rbac.MustLoadAdminConfig()
	rbac.RegisterAdmin(public, protected, db, binder, authCfg, rdb, captchaSvc)

	setting.RegisterAdmin(protected, db, binder)
	settingSvc := provider.NewSettingService(db)
	config.RegisterAdmin(protected, db, binder, settingSvc)
	contact.RegisterAdmin(protected, db, binder)
	article.RegisterAdmin(protected, db, binder)
	user.RegisterAdmin(protected, db, binder)
	region.Register(protected, db, binder)

	upload.RegisterAdmin(protected, db, binder, settingSvc)
	dashboard.RegisterAdmin(protected, db)

	health.Register(r)

	r.LoadHTMLGlob("./views/*.html")
	r.Static("/assets", "./views/assets")
	r.Static("/uploads", "./uploads")
	r.GET("/favicon.ico", func(c *gin.Context) {
		c.File("./views/favicon.ico")
	})
	r.GET("/logo.svg", func(c *gin.Context) {
		c.File("./views/logo.svg")
	})
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})
	r.NoRoute(func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})
	return r
}
