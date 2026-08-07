package runner

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"zero-backend/internal/modules/rbac"

	"github.com/241x/zero-kit/apperror"
	"github.com/241x/zero-kit/logger"
	"github.com/241x/zero-web/errcode"
)

// OpenAPISpec OpenAPI 3.x 规范结构
type OpenAPISpec struct {
	OpenAPI string `json:"openapi"`
	Info    struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Version     string `json:"version"`
	} `json:"info"`
	Paths map[string]PathItem `json:"paths"`
}

// PathItem 路径定义
type PathItem map[string]Operation

// Operation 操作定义
type Operation struct {
	Tags        []string `json:"tags"`
	Summary     string   `json:"summary"`
	Description string   `json:"description"`
	OperationID string   `json:"operationId"`
}

// SyncApiRunner API同步执行器
type SyncApiRunner struct {
	logger logger.Logger
	repo   *rbac.RbacApiRepository
}

// NewSyncApiRunner 创建API同步执行器
func NewSyncApiRunner(l logger.Logger, repo *rbac.RbacApiRepository) *SyncApiRunner {
	return &SyncApiRunner{
		logger: l,
		repo:   repo,
	}
}

// SyncResult 同步结果
type SyncResult struct {
	TotalInDoc int      // 文档中的API数量
	Added      int      // 新增数量
	Updated    int      // 更新数量
	Deleted    int      // 删除数量
	Skipped    int      // 跳过数量
	Errors     []string // 错误列表
}

// Run 执行同步
func (r *SyncApiRunner) Run(ctx context.Context, filePath string, deleteExtra bool) (*SyncResult, error) {
	if filePath == "" {
		filePath = filepath.Join("docs", "admin", "index.json")
	}

	spec, err := r.parseOpenAPI(filePath)
	if err != nil {
		return nil, err
	}

	result := &SyncResult{
		Errors: make([]string, 0),
	}

	tagToParentID := r.ensureTagCategories(ctx, spec, result)
	docURLs := make(map[string]struct{})

	for path, methods := range spec.Paths {
		for method, op := range methods {
			if !strings.EqualFold(method, "POST") {
				continue
			}

			result.TotalInDoc++

			parentID := r.resolveParentID(op, tagToParentID)
			name := r.resolveAPIName(op, path)

			docURLs[path] = struct{}{}

			if err := r.syncSingleAPI(ctx, path, name, parentID, result); err != nil {
				result.Errors = append(result.Errors, err.Error())
			}
		}
	}

	if deleteExtra {
		deleted, err := r.deleteExtraAPIs(ctx, docURLs)
		if err != nil {
			result.Errors = append(result.Errors, "删除多余API失败: "+err.Error())
		} else {
			result.Deleted = deleted
		}
	}

	r.logger.Info("同步完成",
		"文档API数", result.TotalInDoc,
		"新增", result.Added,
		"更新", result.Updated,
		"删除", result.Deleted,
		"跳过", result.Skipped,
		"错误", len(result.Errors),
	)

	return result, nil
}

// ensureTagCategories 确保所有tag对应的父级目录存在
func (r *SyncApiRunner) ensureTagCategories(ctx context.Context, spec *OpenAPISpec, result *SyncResult) map[string]uint32 {
	tagToParentID := make(map[string]uint32)
	uniqueTags := r.collectUniqueTags(spec)

	r.logger.Info("发现分组数量", "count", len(uniqueTags))

	for _, tag := range uniqueTags {
		parentID, err := r.ensureParentCategory(ctx, tag)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("创建父级目录失败 %s: %s", tag, err))
			continue
		}
		tagToParentID[tag] = parentID
	}

	return tagToParentID
}

// resolveAPIName 解析API名称，按优先级：Summary > OperationID > path
func (r *SyncApiRunner) resolveAPIName(op Operation, fallbackPath string) string {
	if op.Summary != "" {
		return op.Summary
	}
	if op.OperationID != "" {
		return op.OperationID
	}
	return fallbackPath
}

// resolveParentID 解析父级目录ID
func (r *SyncApiRunner) resolveParentID(op Operation, tagToParentID map[string]uint32) uint32 {
	if len(op.Tags) > 0 && op.Tags[0] != "" {
		return tagToParentID[op.Tags[0]]
	}
	return 0
}

// syncSingleAPI 同步单个API（新增或更新）
func (r *SyncApiRunner) syncSingleAPI(ctx context.Context, url, name string, parentID uint32, result *SyncResult) error {
	existing, err := r.repo.GetAPIByPath(ctx, url)
	if err != nil {
		return fmt.Errorf("查询API失败 %s: %w", url, err)
	}

	if existing != nil && existing.ID > 0 {
		return r.updateExistingAPI(ctx, existing, name, parentID, url, result)
	}

	if err := r.createAPI(ctx, url, name, parentID); err != nil {
		return fmt.Errorf("创建API失败 %s: %w", url, err)
	}

	result.Added++
	r.logger.Info("新增API", "url", url, "name", name, "parent_id", parentID)
	return nil
}

// updateExistingAPI 更新已存在的API（如需更新）
func (r *SyncApiRunner) updateExistingAPI(ctx context.Context, existing *rbac.RbacApi, name string, parentID uint32, url string, result *SyncResult) error {
	if existing.Name == name && existing.ParentId == parentID {
		result.Skipped++
		return nil
	}

	if err := r.updateAPI(ctx, existing.ID, name, parentID); err != nil {
		return fmt.Errorf("更新API失败 %s: %w", url, err)
	}

	result.Updated++
	r.logger.Info("更新API", "url", url, "name", name, "parent_id", parentID)
	return nil
}

// collectUniqueTags 收集所有唯一的 tag
func (r *SyncApiRunner) collectUniqueTags(spec *OpenAPISpec) []string {
	tagSet := make(map[string]struct{})
	for _, methods := range spec.Paths {
		for _, op := range methods {
			if len(op.Tags) > 0 && op.Tags[0] != "" {
				tagSet[op.Tags[0]] = struct{}{}
			}
		}
	}

	tags := make([]string, 0, len(tagSet))
	for tag := range tagSet {
		tags = append(tags, tag)
	}
	return tags
}

// ensureParentCategory 确保父级目录存在，返回父级目录的ID
func (r *SyncApiRunner) ensureParentCategory(ctx context.Context, tag string) (uint32, error) {
	existing, err := r.repo.GetAPIByName(ctx, tag)
	if err != nil {
		return 0, err
	}

	if existing != nil && existing.ID > 0 {
		return existing.ID, nil
	}

	api := &rbac.RbacApi{
		Name:     tag,
		Url:      rbac.APICategoryMarker,
		ParentId: 0,
		Sort:     100,
	}

	if err := r.repo.Create(ctx, api); err != nil {
		return 0, err
	}

	r.logger.Info("创建父级目录", "name", tag)
	return api.ID, nil
}

// parseOpenAPI 解析OpenAPI文档
func (r *SyncApiRunner) parseOpenAPI(filePath string) (*OpenAPISpec, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, apperror.New(errcode.NotFound, apperror.WithMsg("找不到 OpenAPI 文档文件："+filePath))
		}
		return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("读取OpenAPI文档失败"))
	}

	var spec OpenAPISpec
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("解析OpenAPI文档失败"))
	}

	if spec.OpenAPI == "" {
		return nil, apperror.New(errcode.InvalidInput, apperror.WithMsg("无效的 OpenAPI 文档：缺少 openapi 字段"))
	}

	if len(spec.Paths) == 0 {
		return nil, apperror.New(errcode.InvalidInput, apperror.WithMsg("OpenAPI 文档中未找到任何路径"))
	}

	r.logger.Info("解析 OpenAPI 文档成功", "版本", spec.OpenAPI, "路径数", len(spec.Paths))

	return &spec, nil
}

// createAPI 创建新API
func (r *SyncApiRunner) createAPI(ctx context.Context, url, name string, parentID uint32) error {
	api := &rbac.RbacApi{
		Name:     name,
		Url:      url,
		ParentId: parentID,
		Sort:     100,
	}
	return r.repo.Create(ctx, api)
}

// updateAPI 更新API
func (r *SyncApiRunner) updateAPI(ctx context.Context, id uint32, name string, parentID uint32) error {
	api := &rbac.RbacApi{ID: id}
	return r.repo.Updates(ctx, api, map[string]any{
		"name":      name,
		"parent_id": parentID,
	})
}

// deleteExtraAPIs 删除文档中不存在的API
func (r *SyncApiRunner) deleteExtraAPIs(ctx context.Context, docURLs map[string]struct{}) (int, error) {
	allAPIs, err := r.repo.FindAll(ctx, nil, nil, nil)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, api := range allAPIs {
		if _, exists := docURLs[api.Url]; !exists {
			r.logger.Info("删除多余API", "url", api.Url, "name", api.Name)
			if err := r.repo.Delete(ctx, api.ID); err != nil {
				return count, err
			}
			count++
		}
	}

	return count, nil
}
