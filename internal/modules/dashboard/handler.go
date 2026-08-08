package dashboard

import (
	"github.com/241x/zero-kit/apperror"
	"github.com/241x/zero-web/ctxkeys"
	"github.com/241x/zero-web/errcode"
	"github.com/241x/zero-web/response"
	"github.com/gin-gonic/gin"
)

// handler 仪表盘处理器
type handler struct {
	svc *Service
}

// stats 获取平台仪表盘统计(全局聚合)
func (h *handler) stats(ctx *gin.Context) {
	result, err := h.svc.Stats(ctx.Request.Context())
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, "请求成功", result)
}

// storeStats 获取企业仪表盘统计(限定当前企业)
func (h *handler) storeStats(ctx *gin.Context) {
	sid := ctxkeys.StoreID(ctx.Request.Context())
	if sid == 0 {
		response.Error(ctx, apperror.New(errcode.InvalidInput, apperror.WithMsg("企业标识缺失")))
		return
	}
	result, err := h.svc.StoreStats(ctx.Request.Context(), sid)
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, "请求成功", result)
}
