package contact

import (
	"github.com/241x/zero-kit/bind"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterAdmin 注册通讯录路由（企业端）
func RegisterAdmin(rg *gin.RouterGroup, db *gorm.DB, binder *bind.Binder) {
	departmentRepo := NewWecomDepartmentRepository(db)
	memberRepo := NewWecomMemberRepository(db)
	memberDepartmentRepo := NewWecomMemberDepartmentRepository(db)
	h := newHandler(binder, NewService(departmentRepo, memberRepo, memberDepartmentRepo))
	rg.POST("/wecom/contact/department/tree", h.DepartmentTree)
	rg.POST("/wecom/contact/member/list", h.MemberList)
	rg.POST("/wecom/contact/member/detail", h.MemberDetail)
}
