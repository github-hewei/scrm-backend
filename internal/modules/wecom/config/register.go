package config

import (
	"github.com/241x/zero-kit/bind"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterAdmin 注册企业微信配置路由（企业端）
func RegisterAdmin(rg *gin.RouterGroup, db *gorm.DB, binder *bind.Binder, settings SettingProvider) {
	configRepo := NewWecomConfigRepository(db)
	appRepo := NewWecomAppRepository(db)
	h := newHandler(binder, NewService(configRepo, appRepo, settings))
	rg.POST("/wecom/config/info", h.GetConfig)
	rg.POST("/wecom/config/save", h.SaveConfig)
	rg.POST("/wecom/app/list", h.ListApps)
	rg.POST("/wecom/app/callback-url", h.GetCallbackUrl)
	rg.POST("/wecom/app/save", h.SaveApp)
	rg.POST("/wecom/app/delete", h.DeleteApp)
}
