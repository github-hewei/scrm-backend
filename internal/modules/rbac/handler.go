package rbac

import (
	"github.com/241x/zero-kit/apperror"
	"github.com/241x/zero-kit/bind"
	"github.com/241x/zero-web/ctxkeys"
	"github.com/241x/zero-web/errcode"
	"github.com/241x/zero-web/response"
	"github.com/gin-gonic/gin"
)

// handler rbac 模块处理器
type handler struct {
	binder    *bind.Binder
	authServ  *AuthService
	authCfg   Config
	menuServ  *RbacMenuService
	apiServ   *RbacApiService
	roleServ  *RbacRoleService
	userServ  *RbacUserService
	storeServ *RbacStoreService
	authMid   *AuthMiddleware
}

// newHandler 创建处理器
func newHandler(
	binder *bind.Binder,
	authServ *AuthService,
	authCfg Config,
	menuServ *RbacMenuService,
	apiServ *RbacApiService,
	roleServ *RbacRoleService,
	userServ *RbacUserService,
	storeServ *RbacStoreService,
	authMid *AuthMiddleware,
) *handler {
	return &handler{
		binder:    binder,
		authServ:  authServ,
		authCfg:   authCfg,
		menuServ:  menuServ,
		apiServ:   apiServ,
		roleServ:  roleServ,
		userServ:  userServ,
		storeServ: storeServ,
		authMid:   authMid,
	}
}

// storeID 获取当前请求的店铺ID
func storeID(ctx *gin.Context) uint32 {
	return ctxkeys.StoreID(ctx.Request.Context())
}

// login 系统登录
func (h *handler) login(ctx *gin.Context) {
	req := &AuthLoginRequest{}
	if err := h.binder.ShouldBindJSON(ctx, req); err != nil {
		response.Error(ctx, err)
		return
	}
	result, refreshToken, err := h.authServ.Login(ctx.Request.Context(), req)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	ctx.SetCookie("token", refreshToken, h.authCfg.RefreshTokenTtl, "/api/refresh-token", "", true, true)
	response.Success(ctx, "请求成功", result)
}

// logout 系统登出
func (h *handler) logout(ctx *gin.Context) {
	response.Success(ctx, "操作成功", nil)
}

// refreshToken 刷新令牌
func (h *handler) refreshToken(ctx *gin.Context) {
	cookie, err := ctx.Request.Cookie("token")
	if err != nil {
		response.Error(ctx, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("读取Cookie失败")))
		return
	}
	result, err := h.authServ.RefreshToken(ctx.Request.Context(), cookie.Value)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, "请求成功", result)
}

// changePassword 修改密码
func (h *handler) changePassword(ctx *gin.Context) {
	req := &ChangePasswordRequest{}
	if err := h.binder.ShouldBindJSON(ctx, req); err != nil {
		response.Error(ctx, err)
		return
	}
	if err := h.authServ.ChangePassword(ctx.Request.Context(), req); err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, "密码修改成功", nil)
}

// permissions 获取用户权限菜单
func (h *handler) permissions(ctx *gin.Context) {
	req := &AuthGetPermissionsRequest{}
	if err := h.binder.ShouldBindJSON(ctx, req); err != nil {
		response.Error(ctx, err)
		return
	}
	result, err := h.authServ.GetPermissions(ctx.Request.Context(), req)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, "请求成功", result)
}

// menuList 获取菜单列表
func (h *handler) menuList(ctx *gin.Context) {
	result, err := h.menuServ.FindTreeList(ctx.Request.Context())
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, "请求成功", result)
}

// menuCreate 创建菜单
func (h *handler) menuCreate(ctx *gin.Context) {
	req := &RbacMenuCreateRequest{}
	if err := h.binder.ShouldBindJSON(ctx, req); err != nil {
		response.Error(ctx, err)
		return
	}
	if err := h.menuServ.Create(ctx.Request.Context(), req); err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, "保存成功", nil)
}

// menuUpdate 更新菜单
func (h *handler) menuUpdate(ctx *gin.Context) {
	req := &RbacMenuUpdateRequest{}
	if err := h.binder.ShouldBindJSON(ctx, req); err != nil {
		response.Error(ctx, err)
		return
	}
	if err := h.menuServ.Update(ctx.Request.Context(), req); err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, "保存成功", nil)
}

// menuDelete 删除菜单
func (h *handler) menuDelete(ctx *gin.Context) {
	req := &RbacMenuDeleteRequest{}
	if err := h.binder.ShouldBindJSON(ctx, req); err != nil {
		response.Error(ctx, err)
		return
	}
	if err := h.menuServ.Delete(ctx.Request.Context(), req); err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, "删除成功", nil)
}

// menuSync 同步菜单
func (h *handler) menuSync(ctx *gin.Context) {
	var req []RbacMenuSyncRequest
	if err := h.binder.ShouldBindJSONArray(ctx, &req); err != nil {
		response.Error(ctx, err)
		return
	}
	if err := h.menuServ.Sync(ctx.Request.Context(), req); err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, "同步成功", nil)
}

// menuApiList 获取菜单关联的接口列表
func (h *handler) menuApiList(ctx *gin.Context) {
	req := &RbacMenuApiListRequest{}
	if err := h.binder.ShouldBindJSON(ctx, req); err != nil {
		response.Error(ctx, err)
		return
	}
	result, err := h.menuServ.FindApiList(ctx.Request.Context(), req)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, "获取成功", result)
}

// menuApiSave 保存菜单关联的接口列表
func (h *handler) menuApiSave(ctx *gin.Context) {
	req := &RbacMenuApiSaveRequest{}
	if err := h.binder.ShouldBindJSON(ctx, req); err != nil {
		response.Error(ctx, err)
		return
	}
	if err := h.menuServ.SaveApiList(ctx.Request.Context(), req); err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, "保存成功", nil)
}

// apiList 获取接口列表
func (h *handler) apiList(ctx *gin.Context) {
	result, err := h.apiServ.FindTreeList(ctx.Request.Context())
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, "请求成功", result)
}

// apiCreate 创建接口
func (h *handler) apiCreate(ctx *gin.Context) {
	req := &RbacApiCreateRequest{}
	if err := h.binder.ShouldBindJSON(ctx, req); err != nil {
		response.Error(ctx, err)
		return
	}
	if err := h.apiServ.Create(ctx.Request.Context(), req); err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, "保存成功", nil)
}

// apiUpdate 更新接口
func (h *handler) apiUpdate(ctx *gin.Context) {
	req := &RbacApiUpdateRequest{}
	if err := h.binder.ShouldBindJSON(ctx, req); err != nil {
		response.Error(ctx, err)
		return
	}
	if err := h.apiServ.Update(ctx.Request.Context(), req); err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, "保存成功", nil)
}

// apiSync 同步接口
func (h *handler) apiSync(ctx *gin.Context) {
	req := []RbacApiSyncRequest{}
	if err := h.binder.ShouldBindJSONArray(ctx, &req); err != nil {
		response.Error(ctx, err)
		return
	}
	if err := h.apiServ.Sync(ctx.Request.Context(), req); err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, "同步成功", nil)
}

// apiDelete 删除接口
func (h *handler) apiDelete(ctx *gin.Context) {
	req := &RbacApiDeleteRequest{}
	if err := h.binder.ShouldBindJSON(ctx, req); err != nil {
		response.Error(ctx, err)
		return
	}
	if err := h.apiServ.Delete(ctx.Request.Context(), req); err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, "删除成功", nil)
}

// storeList 获取企业列表
func (h *handler) storeList(ctx *gin.Context) {
	req := &RbacStoreListRequest{}
	if err := h.binder.ShouldBindJSON(ctx, req); err != nil {
		response.Error(ctx, err)
		return
	}
	result, err := h.storeServ.FindList(ctx.Request.Context(), req)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, "请求成功", result)
}

// storeCreate 创建企业
func (h *handler) storeCreate(ctx *gin.Context) {
	req := &RbacStoreCreateRequest{}
	if err := h.binder.ShouldBindJSON(ctx, req); err != nil {
		response.Error(ctx, err)
		return
	}
	result, err := h.storeServ.Create(ctx.Request.Context(), req)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, "创建成功", result)
}

// storeUpdate 更新企业
func (h *handler) storeUpdate(ctx *gin.Context) {
	req := &RbacStoreUpdateRequest{}
	if err := h.binder.ShouldBindJSON(ctx, req); err != nil {
		response.Error(ctx, err)
		return
	}
	if err := h.storeServ.Update(ctx.Request.Context(), req); err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, "更新成功", nil)
}

// storeDelete 删除企业
func (h *handler) storeDelete(ctx *gin.Context) {
	req := &RbacStoreDeleteRequest{}
	if err := h.binder.ShouldBindJSON(ctx, req); err != nil {
		response.Error(ctx, err)
		return
	}
	if err := h.storeServ.Delete(ctx.Request.Context(), req); err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, "删除成功", nil)
}

// storeRecycle 回收企业
func (h *handler) storeRecycle(ctx *gin.Context) {
	req := &RbacStoreDeleteRequest{}
	if err := h.binder.ShouldBindJSON(ctx, req); err != nil {
		response.Error(ctx, err)
		return
	}
	if err := h.storeServ.Recycle(ctx.Request.Context(), req); err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, "操作成功", nil)
}

// storeRestore 恢复企业
func (h *handler) storeRestore(ctx *gin.Context) {
	req := &RbacStoreDeleteRequest{}
	if err := h.binder.ShouldBindJSON(ctx, req); err != nil {
		response.Error(ctx, err)
		return
	}
	if err := h.storeServ.Restore(ctx.Request.Context(), req); err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, "操作成功", nil)
}

// roleList 获取角色列表
func (h *handler) roleList(ctx *gin.Context) {
	req := &RbacRoleListRequest{}
	if err := h.binder.ShouldBindJSON(ctx, req); err != nil {
		response.Error(ctx, err)
		return
	}
	if sid := storeID(ctx); sid != 0 {
		req.StoreId = sid
		req.IsSuper = 0
	} else {
		req.IsSuper = 1
	}
	result, err := h.roleServ.FindAll(ctx.Request.Context(), req)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, "请求成功", result)
}

// platformRoleList 平台端获取角色列表(分页，仅超管角色，可按企业过滤)
func (h *handler) platformRoleList(ctx *gin.Context) {
	req := &RbacRolePageListRequest{}
	if err := h.binder.ShouldBindJSON(ctx, req); err != nil {
		response.Error(ctx, err)
		return
	}
	req.IsSuper = 1
	result, err := h.roleServ.FindList(ctx.Request.Context(), req)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, "请求成功", result)
}

// roleCreate 创建角色
func (h *handler) roleCreate(ctx *gin.Context) {
	req := &RbacRoleCreateRequest{}
	if err := h.binder.ShouldBindJSON(ctx, req); err != nil {
		response.Error(ctx, err)
		return
	}
	if sid := storeID(ctx); sid != 0 {
		req.StoreId = sid
		req.IsSuper = 0
	} else {
		req.IsSuper = 1
	}
	if err := h.roleServ.Create(ctx.Request.Context(), req); err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, "创建成功", nil)
}

// roleUpdate 更新角色
func (h *handler) roleUpdate(ctx *gin.Context) {
	req := &RbacRoleUpdateRequest{}
	if err := h.binder.ShouldBindJSON(ctx, req); err != nil {
		response.Error(ctx, err)
		return
	}
	if sid := storeID(ctx); sid != 0 {
		req.StoreId = sid
		req.IsSuper = 0
	} else {
		req.IsSuper = 1
	}
	if err := h.roleServ.Update(ctx.Request.Context(), req); err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, "更新成功", nil)
}

// roleDelete 删除角色
func (h *handler) roleDelete(ctx *gin.Context) {
	req := &RbacRoleDeleteRequest{}
	if err := h.binder.ShouldBindJSON(ctx, req); err != nil {
		response.Error(ctx, err)
		return
	}
	if sid := storeID(ctx); sid != 0 {
		req.StoreId = sid
	}
	if err := h.roleServ.Delete(ctx.Request.Context(), req); err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, "删除成功", nil)
}

// roleSetMenus 设置角色菜单
func (h *handler) roleSetMenus(ctx *gin.Context) {
	req := &RbacRoleMenuSetRequest{}
	if err := h.binder.ShouldBindJSON(ctx, req); err != nil {
		response.Error(ctx, err)
		return
	}
	if sid := storeID(ctx); sid != 0 {
		req.StoreId = sid
	}
	if err := h.roleServ.SetMenus(ctx.Request.Context(), req); err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, "设置成功", nil)
}

// userList 获取后台用户列表
func (h *handler) userList(ctx *gin.Context) {
	req := &RbacUserListRequest{}
	if err := h.binder.ShouldBindJSON(ctx, req); err != nil {
		response.Error(ctx, err)
		return
	}
	if sid := storeID(ctx); sid != 0 {
		req.StoreId = sid
		req.IsSuper = 0
	} else {
		req.IsSuper = 1
	}
	result, err := h.userServ.FindList(ctx.Request.Context(), req)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, "请求成功", result)
}

// userCreate 创建后台用户
func (h *handler) userCreate(ctx *gin.Context) {
	req := &RbacUserCreateRequest{}
	if err := h.binder.ShouldBindJSON(ctx, req); err != nil {
		response.Error(ctx, err)
		return
	}
	if sid := storeID(ctx); sid != 0 {
		req.StoreId = sid
		req.IsSuper = 0
	} else {
		req.IsSuper = 1
	}
	if err := h.userServ.Create(ctx.Request.Context(), req); err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, "创建成功", nil)
}

// userUpdate 更新后台用户
func (h *handler) userUpdate(ctx *gin.Context) {
	req := &RbacUserUpdateRequest{}
	if err := h.binder.ShouldBindJSON(ctx, req); err != nil {
		response.Error(ctx, err)
		return
	}
	if sid := storeID(ctx); sid != 0 {
		req.StoreId = sid
		req.IsSuper = 0
	} else {
		req.IsSuper = 1
	}
	if err := h.userServ.Update(ctx.Request.Context(), req); err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, "更新成功", nil)
}

// userDelete 删除后台用户
func (h *handler) userDelete(ctx *gin.Context) {
	req := &RbacUserDeleteRequest{}
	if err := h.binder.ShouldBindJSON(ctx, req); err != nil {
		response.Error(ctx, err)
		return
	}
	if sid := storeID(ctx); sid != 0 {
		req.StoreId = sid
	}
	if err := h.userServ.Delete(ctx.Request.Context(), req); err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, "删除成功", nil)
}

// userSetRoles 设置用户角色
func (h *handler) userSetRoles(ctx *gin.Context) {
	req := &RbacUserRoleSetRequest{}
	if err := h.binder.ShouldBindJSON(ctx, req); err != nil {
		response.Error(ctx, err)
		return
	}
	if sid := storeID(ctx); sid != 0 {
		req.StoreId = sid
	}
	if err := h.userServ.SetRoles(ctx.Request.Context(), req); err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, "设置成功", nil)
}

// userResetPassword 重置用户密码
func (h *handler) userResetPassword(ctx *gin.Context) {
	req := &RbacUserResetPasswordRequest{}
	if err := h.binder.ShouldBindJSON(ctx, req); err != nil {
		response.Error(ctx, err)
		return
	}
	if sid := storeID(ctx); sid != 0 {
		req.StoreId = sid
	}
	newPassword, err := h.userServ.ResetPassword(ctx.Request.Context(), req)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, "重置成功", gin.H{
		"new_password": newPassword,
	})
}
