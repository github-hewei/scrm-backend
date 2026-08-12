package config

import (
	"github.com/241x/zero-kit/apperror"
	"github.com/241x/zero-kit/bind"
	"github.com/241x/zero-web/ctxkeys"
	"github.com/241x/zero-web/errcode"
	"github.com/241x/zero-web/response"
	"github.com/gin-gonic/gin"
)

// Handler 企业微信配置处理器
type Handler struct {
	binder *bind.Binder
	svc    *Service
}

// newHandler 创建企业微信配置处理器
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

// GetConfig 获取企业微信配置
func (h *Handler) GetConfig(c *gin.Context) {
	storeId, ok := h.requireStoreId(c)
	if !ok {
		return
	}
	result, err := h.svc.GetConfig(c.Request.Context(), storeId)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, "请求成功", result)
}

// SaveConfig 保存企业微信配置
func (h *Handler) SaveConfig(c *gin.Context) {
	req := &ConfigSaveRequest{}
	if err := h.binder.ShouldBindJSON(c, req); err != nil {
		response.Error(c, err)
		return
	}
	storeId, ok := h.requireStoreId(c)
	if !ok {
		return
	}
	if err := h.svc.SaveConfig(c.Request.Context(), storeId, req); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, "保存成功", nil)
}

// GetCallbackUrl 生成回调地址
func (h *Handler) GetCallbackUrl(c *gin.Context) {
	storeId, ok := h.requireStoreId(c)
	if !ok {
		return
	}
	result, err := h.svc.GetCallbackUrl(c.Request.Context(), storeId)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, "请求成功", result)
}

// ListApps 获取应用列表
func (h *Handler) ListApps(c *gin.Context) {
	storeId, ok := h.requireStoreId(c)
	if !ok {
		return
	}
	result, err := h.svc.ListApps(c.Request.Context(), storeId)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, "请求成功", result)
}

// SaveApp 保存应用凭据
func (h *Handler) SaveApp(c *gin.Context) {
	req := &AppSaveRequest{}
	if err := h.binder.ShouldBindJSON(c, req); err != nil {
		response.Error(c, err)
		return
	}
	storeId, ok := h.requireStoreId(c)
	if !ok {
		return
	}
	result, err := h.svc.SaveApp(c.Request.Context(), storeId, req)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, "保存成功", result)
}

// DeleteApp 删除应用凭据
func (h *Handler) DeleteApp(c *gin.Context) {
	req := &AppDeleteRequest{}
	if err := h.binder.ShouldBindJSON(c, req); err != nil {
		response.Error(c, err)
		return
	}
	storeId, ok := h.requireStoreId(c)
	if !ok {
		return
	}
	if err := h.svc.DeleteApp(c.Request.Context(), storeId, req.ID); err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, "删除成功", nil)
}
