package sync

import (
	"zero-backend/internal/modules/async"
	wecomconfig "zero-backend/internal/modules/wecom/config"

	"github.com/241x/zero-kit/bind"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterAdmin 注册企业微信同步任务扩展点（企业端）。
// 任务的触发/列表/详情统一走 async 模块的通用接口（/async/task/submit|list|detail）
func RegisterAdmin(rg *gin.RouterGroup, db *gorm.DB, binder *bind.Binder) {
	async.RegisterSubmitHandler(JobTypeWecomSync, NewSubmitHandler(wecomconfig.NewWecomConfigRepository(db)))
}
