package sync

import (
	wecomconfig "zero-backend/internal/modules/wecom/config"

	"github.com/241x/zero-kit/bind"
	"github.com/241x/zero-kit/job"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterAdmin 注册企业微信同步作业路由（企业端）。
// jobs 表由 job.SQLStore AutoMigrate 自动创建（幂等）
func RegisterAdmin(rg *gin.RouterGroup, db *gorm.DB, binder *bind.Binder) {
	configRepo := wecomconfig.NewWecomConfigRepository(db)
	store, err := job.NewSQLStore(db)
	if err != nil {
		panic(err)
	}
	h := newHandler(binder, NewJobService(configRepo, store))
	rg.POST("/wecom/sync/trigger", h.Submit)
	rg.POST("/wecom/sync/jobs", h.List)
	rg.POST("/wecom/sync/job/detail", h.Detail)
}
