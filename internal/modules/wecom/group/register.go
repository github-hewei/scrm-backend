package group

import (
	"github.com/241x/zero-kit/bind"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterAdmin 注册客户群路由（企业端）
func RegisterAdmin(rg *gin.RouterGroup, db *gorm.DB, binder *bind.Binder) {
	groupRepo := NewWecomGroupRepository(db)
	groupMemberRepo := NewWecomGroupMemberRepository(db)
	h := newHandler(binder, NewService(groupRepo, groupMemberRepo))
	rg.POST("/wecom/group/list", h.List)
	rg.POST("/wecom/group/detail", h.Detail)
}
