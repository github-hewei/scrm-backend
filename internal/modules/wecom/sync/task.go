package sync

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"

	"zero-backend/internal/modules/async"
	wecomconfig "zero-backend/internal/modules/wecom/config"

	"github.com/241x/zero-kit/apperror"
	"github.com/241x/zero-kit/baserepo"
	"github.com/241x/zero-web/errcode"
)

// JobTypeWecomSync 企业微信数据同步作业类型（worker 处理器按此注册）
const JobTypeWecomSync = "wecom.sync"

// WecomSyncJobPayload 同步作业入参
type WecomSyncJobPayload struct {
	StoreId uint32 `json:"store_id"` // 企业ID，0=全部已接入企业
	Scope   string `json:"scope"`    // all|dept|contact|group
}

// NewSubmitHandler 构建同步任务的提交规则，注册到 async 通用提交接口：
// 校验企业接入与 scope、生成租户标题、按企业去重（连点保护）
func NewSubmitHandler(configRepo *wecomconfig.WecomConfigRepository) async.SubmitHandler {
	return async.SubmitHandler{
		Validate: func(ctx context.Context, storeId uint32, payload []byte) error {
			var req WecomSyncJobPayload
			if err := json.Unmarshal(payload, &req); err != nil {
				return apperror.New(errcode.InvalidInput, apperror.WithMsg("同步任务参数解析失败"))
			}
			parsed, err := ParseScope(req.Scope)
			if err != nil {
				return apperror.New(errcode.InvalidInput, apperror.WithMsg(err.Error()))
			}
			if !IsSupported(parsed) {
				return apperror.New(errcode.InvalidInput, apperror.WithMsgf("暂不支持同步范围: %s", req.Scope))
			}
			if _, err := configRepo.FindOne(ctx, &wecomconfig.WecomConfigFilter{StoreId: storeId, Status: 1}); err != nil {
				if errors.Is(err, baserepo.ErrRecordNotFound) {
					return apperror.New(errcode.NotFound, apperror.WithMsgf("企业未接入或已停用 store_id=%d", storeId))
				}
				return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("校验企业接入配置失败"))
			}
			return nil
		},
		Title: func(payload []byte) string {
			var req WecomSyncJobPayload
			_ = json.Unmarshal(payload, &req)
			return titleForScope(req.Scope)
		},
		DedupKey: func(payload []byte) string {
			var req WecomSyncJobPayload
			_ = json.Unmarshal(payload, &req)
			return strconv.FormatUint(uint64(req.StoreId), 10)
		},
	}
}

// titleForScope 生成租户友好的任务标题
func titleForScope(scope string) string {
	switch scope {
	case string(ScopeGroup):
		return "同步企业微信客户群"
	case string(ScopeContact):
		return "同步企业微信客户"
	case string(ScopeDept):
		return "同步企业微信通讯录"
	case string(ScopeAll):
		return "同步企业微信全量数据"
	}
	return "企业微信数据同步"
}
