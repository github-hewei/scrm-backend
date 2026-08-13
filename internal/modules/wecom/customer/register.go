package customer

import (
	"github.com/241x/zero-kit/bind"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterAdmin 注册客户路由（企业端）。memberNames 为跨包成员姓名查询实现
func RegisterAdmin(rg *gin.RouterGroup, db *gorm.DB, binder *bind.Binder, memberNames MemberNameProvider) {
	customerRepo := NewWecomCustomerRepository(db)
	followRepo := NewWecomCustomerFollowRepository(db)
	h := newHandler(binder, NewService(customerRepo, followRepo, memberNames))
	rg.POST("/wecom/customer/list", h.List)
	rg.POST("/wecom/customer/detail", h.Detail)
}
