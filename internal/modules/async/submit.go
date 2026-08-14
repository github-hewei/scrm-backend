package async

import (
	"context"
	"sync"
)

// SubmitHandler 任务类型扩展点：业务侧注册提交校验/标题/去重规则，async 保持零业务依赖。
// 注册发生在服务启动（路由组装）阶段，运行期只读
type SubmitHandler struct {
	// Validate 提交前业务校验（如企业是否接入），返回的错误透传给调用方；nil 表示不校验
	Validate func(ctx context.Context, storeId uint32, payload []byte) error
	// Title 生成租户友好标题；nil 时标题为空
	Title func(payload []byte) string
	// DedupKey 提取去重键（如企业ID）；返回空串表示不去重
	DedupKey func(payload []byte) string
	// MaxAttempts 最大执行次数（含首次，失败自动重试），0 使用服务默认值
	MaxAttempts int
}

var (
	submitMu       sync.RWMutex
	submitHandlers = make(map[string]SubmitHandler)
)

// RegisterSubmitHandler 注册任务类型的提交规则，需在服务启动（路由组装）时完成
func RegisterSubmitHandler(taskType string, h SubmitHandler) {
	submitMu.Lock()
	defer submitMu.Unlock()
	submitHandlers[taskType] = h
}

// lookupSubmitHandler 查询任务类型的提交规则
func lookupSubmitHandler(taskType string) (SubmitHandler, bool) {
	submitMu.RLock()
	defer submitMu.RUnlock()
	h, ok := submitHandlers[taskType]
	return h, ok
}
