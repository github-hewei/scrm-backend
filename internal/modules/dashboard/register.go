package dashboard

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterPlatform 注册平台端仪表盘路由
func RegisterPlatform(rg *gin.RouterGroup, db *gorm.DB) {
	svc := NewService(NewRepository(db))
	h := &handler{svc: svc}
	rg.POST("/dashboard/stats", h.stats)
}

// RegisterAdmin 注册企业端仪表盘路由
func RegisterAdmin(rg *gin.RouterGroup, db *gorm.DB) {
	svc := NewService(NewRepository(db))
	h := &handler{svc: svc}
	rg.POST("/dashboard/stats", h.storeStats)
}
