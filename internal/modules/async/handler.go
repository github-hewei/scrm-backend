package async

import (
	"github.com/241x/zero-kit/apperror"
	"github.com/241x/zero-kit/bind"
	"github.com/241x/zero-web/ctxkeys"
	"github.com/241x/zero-web/errcode"
	"github.com/241x/zero-web/response"
	"github.com/gin-gonic/gin"
)

// Handler 异步任务处理器
type Handler struct {
	binder *bind.Binder
	svc    *TaskService
}

// newHandler 创建异步任务处理器
func newHandler(binder *bind.Binder, svc *TaskService) *Handler {
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

// Submit 提交异步任务（仅当前登录企业，任务类型需已注册）
func (h *Handler) Submit(c *gin.Context) {
	req := &SubmitTaskRequest{}
	if err := h.binder.ShouldBindJSON(c, req); err != nil {
		response.Error(c, err)
		return
	}
	storeId, ok := h.requireStoreId(c)
	if !ok {
		return
	}
	jobID, err := h.svc.Submit(c.Request.Context(), storeId, req.TaskType, req.Payload)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, "请求成功", &SubmitTaskResponse{JobId: jobID})
}

// List 任务列表（仅当前登录企业）
func (h *Handler) List(c *gin.Context) {
	req := &TaskListRequest{}
	if err := h.binder.ShouldBindJSON(c, req); err != nil {
		response.Error(c, err)
		return
	}
	storeId, ok := h.requireStoreId(c)
	if !ok {
		return
	}
	list, total, err := h.svc.ListByStore(c.Request.Context(), storeId, req.TaskType, req.Page, req.Limit)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, "请求成功", &TaskListResponse{List: fromTasks(list), Total: total})
}

// Detail 任务详情（仅当前登录企业）
func (h *Handler) Detail(c *gin.Context) {
	req := &TaskDetailRequest{}
	if err := h.binder.ShouldBindJSON(c, req); err != nil {
		response.Error(c, err)
		return
	}
	storeId, ok := h.requireStoreId(c)
	if !ok {
		return
	}
	task, err := h.svc.GetByStore(c.Request.Context(), storeId, req.Id)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, "请求成功", fromTask(task))
}
