package upload

import (
	"strconv"

	"github.com/241x/zero-kit/apperror"
	"github.com/241x/zero-kit/bind"
	"github.com/241x/zero-web/ctxkeys"
	"github.com/241x/zero-web/errcode"
	"github.com/241x/zero-web/response"
	"github.com/gin-gonic/gin"
)

// Handler 上传模块处理器
type Handler struct {
	binder   *bind.Binder
	groupSvc *GroupService
	fileSvc  *FileService
}

// newHandler 创建上传模块处理器
func newHandler(binder *bind.Binder, groupSvc *GroupService, fileSvc *FileService) *Handler {
	return &Handler{binder: binder, groupSvc: groupSvc, fileSvc: fileSvc}
}

// ListGroup 获取上传分组列表
func (h *Handler) ListGroup(ctx *gin.Context) {
	storeId := ctxkeys.StoreID(ctx.Request.Context())
	result, err := h.groupSvc.FindTreeList(ctx.Request.Context(), storeId)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, "请求成功", result)
}

// CreateGroup 创建上传分组
func (h *Handler) CreateGroup(ctx *gin.Context) {
	req := &GroupCreateRequest{}
	if err := h.binder.ShouldBindJSON(ctx, req); err != nil {
		response.Error(ctx, err)
		return
	}
	req.StoreId = ctxkeys.StoreID(ctx.Request.Context())
	if err := h.groupSvc.Create(ctx.Request.Context(), req); err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, "创建成功", nil)
}

// UpdateGroup 更新上传分组
func (h *Handler) UpdateGroup(ctx *gin.Context) {
	req := &GroupUpdateRequest{}
	if err := h.binder.ShouldBindJSON(ctx, req); err != nil {
		response.Error(ctx, err)
		return
	}
	req.StoreId = ctxkeys.StoreID(ctx.Request.Context())
	if err := h.groupSvc.Update(ctx.Request.Context(), req); err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, "更新成功", nil)
}

// DeleteGroup 删除上传分组
func (h *Handler) DeleteGroup(ctx *gin.Context) {
	req := &GroupDeleteRequest{}
	if err := h.binder.ShouldBindJSON(ctx, req); err != nil {
		response.Error(ctx, err)
		return
	}
	req.StoreId = ctxkeys.StoreID(ctx.Request.Context())
	if err := h.groupSvc.Delete(ctx.Request.Context(), req); err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, "删除成功", nil)
}

// ListFile 获取文件列表
func (h *Handler) ListFile(ctx *gin.Context) {
	req := &FileListRequest{}
	if err := h.binder.ShouldBindJSON(ctx, req); err != nil {
		response.Error(ctx, err)
		return
	}
	req.StoreId = ctxkeys.StoreID(ctx.Request.Context())
	result, err := h.fileSvc.FindList(ctx.Request.Context(), req)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, "请求成功", result)
}

// UploadFile 上传文件
func (h *Handler) UploadFile(ctx *gin.Context) {
	file, err := ctx.FormFile("file")
	if err != nil {
		response.Error(ctx, apperror.New(errcode.InvalidInput, apperror.WithMsg("请选择上传文件")))
		return
	}
	groupId, _ := strconv.Atoi(ctx.PostForm("group_id"))
	req := &FileRequest{
		File:    file,
		GroupId: uint32(groupId),
		StoreId: ctxkeys.StoreID(ctx.Request.Context()),
	}
	result, err := h.fileSvc.Upload(ctx.Request.Context(), req)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, "上传成功", result)
}

// UploadFileAdmin 上传文件(管理端)
func (h *Handler) UploadFileAdmin(ctx *gin.Context) {
	file, err := ctx.FormFile("file")
	if err != nil {
		response.Error(ctx, apperror.New(errcode.InvalidInput, apperror.WithMsg("请选择上传文件")))
		return
	}
	groupId, _ := strconv.Atoi(ctx.PostForm("group_id"))
	req := &FileRequest{
		File:    file,
		GroupId: uint32(groupId),
		StoreId: ctxkeys.StoreID(ctx.Request.Context()),
	}
	result, err := h.fileSvc.Upload(ctx.Request.Context(), req)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, "上传成功", result)
}

// DeleteFile 删除文件
func (h *Handler) DeleteFile(ctx *gin.Context) {
	req := &FileDeleteRequest{}
	if err := h.binder.ShouldBindJSON(ctx, req); err != nil {
		response.Error(ctx, err)
		return
	}
	req.StoreId = ctxkeys.StoreID(ctx.Request.Context())
	if err := h.fileSvc.Delete(ctx.Request.Context(), req); err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, "删除成功", nil)
}

// DetailFile 获取文件详情
func (h *Handler) DetailFile(ctx *gin.Context) {
	req := &FileDetailRequest{}
	if err := h.binder.ShouldBindJSON(ctx, req); err != nil {
		response.Error(ctx, err)
		return
	}
	req.StoreId = ctxkeys.StoreID(ctx.Request.Context())
	result, err := h.fileSvc.Detail(ctx.Request.Context(), req)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, "请求成功", result)
}
