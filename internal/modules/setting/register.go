package setting

import (
	"github.com/241x/zero-kit/bind"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// buildHandler 创建设置模块处理器
func buildHandler(db *gorm.DB, binder *bind.Binder) *Handler {
	repo := NewRepository(db)
	defaultRepo := NewDefaultRepository(db)
	svc := NewService(repo, defaultRepo)
	defaultSvc := NewDefaultService(defaultRepo)
	return newHandler(binder, svc, defaultSvc)
}

// RegisterAdmin 注册设置模块路由（管理端）。仅暴露租户级设置和工具接口，不暴露系统级默认设置。
func RegisterAdmin(rg *gin.RouterGroup, db *gorm.DB, binder *bind.Binder) {
	h := buildHandler(db, binder)
	rg.POST("/setting/list", h.List)
	rg.POST("/setting/create", h.Create)
	rg.POST("/setting/update", h.Update)
	rg.POST("/setting/delete", h.Delete)
	rg.POST("/setting/form-configs", h.FormConfigs)
}

// RegisterPlatform 注册设置模块路由（平台端）。暴露系统级默认设置和工具接口。
func RegisterPlatform(rg *gin.RouterGroup, db *gorm.DB, binder *bind.Binder) {
	h := buildHandler(db, binder)
	rg.POST("/setting/default/list", h.DefaultList)
	rg.POST("/setting/default/create", h.DefaultCreate)
	rg.POST("/setting/default/update", h.DefaultUpdate)
	rg.POST("/setting/default/delete", h.DefaultDelete)
	rg.POST("/setting/form-configs", h.FormConfigs)
}

// RegisterApi 注册设置模块路由（API端）
func RegisterApi(rg *gin.RouterGroup, db *gorm.DB, binder *bind.Binder) {
}
