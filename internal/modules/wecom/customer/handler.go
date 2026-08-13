package customer

import (
	"github.com/241x/zero-kit/apperror"
	"github.com/241x/zero-kit/bind"
	"github.com/241x/zero-web/ctxkeys"
	"github.com/241x/zero-web/errcode"
	"github.com/241x/zero-web/response"
	"github.com/gin-gonic/gin"
)

// Handler 客户处理器
type Handler struct {
	binder *bind.Binder
	svc    *Service
}

// newHandler 创建客户处理器
func newHandler(binder *bind.Binder, svc *Service) *Handler {
	return &Handler{binder: binder, svc: svc}
}

// requireStoreId 从上下文获取企业ID，缺失则响应错误并返回 false
func (h *Handler) requireStoreId(c *gin.Context) (uint32, bool) {
	storeId := ctxkeys.StoreID(c.Request.Context())
	if storeId == 0 {
		response.Error(c, apperror.New(errcode.InvalidInput, apperror.WithMsg("企业标识缺失")))
		return 0, false
	}
	return storeId, true
}

// List 客户列表
func (h *Handler) List(c *gin.Context) {
	req := &CustomerListRequest{}
	if err := h.binder.ShouldBindJSON(c, req); err != nil {
		response.Error(c, err)
		return
	}
	storeId, ok := h.requireStoreId(c)
	if !ok {
		return
	}
	result, err := h.svc.FindList(c.Request.Context(), storeId, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, "请求成功", result)
}

// Detail 客户详情
func (h *Handler) Detail(c *gin.Context) {
	req := &CustomerDetailRequest{}
	if err := h.binder.ShouldBindJSON(c, req); err != nil {
		response.Error(c, err)
		return
	}
	storeId, ok := h.requireStoreId(c)
	if !ok {
		return
	}
	result, err := h.svc.FindDetail(c.Request.Context(), storeId, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, "请求成功", result)
}
