package async

import (
	"github.com/241x/zero-kit/bind"
	"github.com/241x/zero-kit/job"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterAdmin 注册通用异步任务路由（企业端）。
// jobs 表由 job.SQLStore AutoMigrate 自动创建（幂等）
func RegisterAdmin(rg *gin.RouterGroup, db *gorm.DB, binder *bind.Binder) {
	store, err := job.NewSQLStore(db)
	if err != nil {
		panic(err)
	}
	svc := NewTaskService(NewAsyncTaskRepository(db), store)
	h := newHandler(binder, svc)
	rg.POST("/async/task/submit", h.Submit)
	rg.POST("/async/task/list", h.List)
	rg.POST("/async/task/detail", h.Detail)
}
