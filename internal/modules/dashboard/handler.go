package dashboard

import (
	"github.com/241x/zero-web/response"
	"github.com/gin-gonic/gin"
)

// handler 仪表盘处理器
type handler struct {
	svc *Service
}

// stats 获取平台仪表盘统计数据
func (h *handler) stats(ctx *gin.Context) {
	result, err := h.svc.Stats(ctx.Request.Context())
	if err != nil {
		response.Error(ctx, err)
		return
	}
	response.Success(ctx, "请求成功", result)
}
