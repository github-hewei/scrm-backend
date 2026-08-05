package upload

import (
	"github.com/241x/zero-kit/bind"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// buildHandler 创建上传模块处理器
func buildHandler(db *gorm.DB, binder *bind.Binder, settings SettingProvider) *Handler {
	groupRepo := NewGroupRepository(db)
	groupSvc := NewGroupService(groupRepo)
	fileRepo := NewFileRepository(db)
	fileSvc := NewFileService(fileRepo, settings)
	return newHandler(binder, groupSvc, fileSvc)
}

// RegisterAdmin 注册上传模块路由（管理端）
func RegisterAdmin(rg *gin.RouterGroup, db *gorm.DB, binder *bind.Binder, settings SettingProvider) {
	h := buildHandler(db, binder, settings)
	rg.POST("/upload/group/list", h.ListGroup)
	rg.POST("/upload/group/create", h.CreateGroup)
	rg.POST("/upload/group/update", h.UpdateGroup)
	rg.POST("/upload/group/delete", h.DeleteGroup)
	rg.POST("/upload/file/list", h.ListFile)
	rg.POST("/upload/file/upload", h.UploadFileAdmin)
	rg.POST("/upload/file/detail", h.DetailFile)
	rg.POST("/upload/file/delete", h.DeleteFile)
}

// RegisterApi 注册上传模块路由（API端）
func RegisterApi(rg *gin.RouterGroup, db *gorm.DB, binder *bind.Binder, settings SettingProvider) {
	h := buildHandler(db, binder, settings)
	rg.POST("/upload/file/list", h.ListFile)
	rg.POST("/upload/file/upload", h.UploadFile)
	rg.POST("/upload/file/detail", h.DetailFile)
	rg.POST("/upload/file/delete", h.DeleteFile)
}

// RegisterPlatform 注册上传模块路由（平台端）
func RegisterPlatform(rg *gin.RouterGroup, db *gorm.DB, binder *bind.Binder, settings SettingProvider) {
	h := buildHandler(db, binder, settings)
	rg.POST("/upload/group/list", h.ListGroup)
	rg.POST("/upload/group/create", h.CreateGroup)
	rg.POST("/upload/group/update", h.UpdateGroup)
	rg.POST("/upload/group/delete", h.DeleteGroup)
	rg.POST("/upload/file/list", h.ListFile)
	rg.POST("/upload/file/upload", h.UploadFileAdmin)
	rg.POST("/upload/file/detail", h.DetailFile)
	rg.POST("/upload/file/delete", h.DeleteFile)
}
