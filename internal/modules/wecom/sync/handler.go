package sync

import (
	"github.com/241x/zero-kit/apperror"
	"github.com/241x/zero-kit/bind"
	"github.com/241x/zero-web/ctxkeys"
	"github.com/241x/zero-web/errcode"
	"github.com/241x/zero-web/response"
	"github.com/gin-gonic/gin"
)

// Handler 同步作业处理器
type Handler struct {
	binder *bind.Binder
	svc    *JobService
}

// newHandler 创建同步作业处理器
func newHandler(binder *bind.Binder, svc *JobService) *Handler {
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

// Submit 提交同步作业（仅限当前登录企业）
func (h *Handler) Submit(c *gin.Context) {
	req := &SubmitSyncRequest{}
	if err := h.binder.ShouldBindJSON(c, req); err != nil {
		response.Error(c, err)
		return
	}
	storeId, ok := h.requireStoreId(c)
	if !ok {
		return
	}
	jobID, err := h.svc.Submit(c.Request.Context(), storeId, req.Scope)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, "请求成功", &SubmitSyncResponse{JobId: jobID})
}

// List 同步作业列表（仅当前登录企业）
func (h *Handler) List(c *gin.Context) {
	req := &JobListRequest{}
	if err := h.binder.ShouldBindJSON(c, req); err != nil {
		response.Error(c, err)
		return
	}
	storeId, ok := h.requireStoreId(c)
	if !ok {
		return
	}
	jobs, total, err := h.svc.List(c.Request.Context(), storeId, req.Page, req.Limit)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, "请求成功", &JobListResponse{List: fromJobs(jobs), Total: total})
}

// Detail 同步作业详情（仅当前登录企业）
func (h *Handler) Detail(c *gin.Context) {
	req := &JobDetailRequest{}
	if err := h.binder.ShouldBindJSON(c, req); err != nil {
		response.Error(c, err)
		return
	}
	storeId, ok := h.requireStoreId(c)
	if !ok {
		return
	}
	j, err := h.svc.Get(c.Request.Context(), storeId, req.JobId)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, "请求成功", fromJob(j))
}
