package upload

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/241x/zero-kit/apperror"
	"github.com/241x/zero-kit/baserepo"
	"github.com/241x/zero-web/errcode"
)

// SettingProvider 设置值读取接口，由宿主项目注入实现。
type SettingProvider interface {
	GetSettingValue(ctx context.Context, key string, storeId uint32, target any) error
}

// GroupService 文件分组服务
type GroupService struct {
	repo *GroupRepository
}

// NewGroupService 创建文件分组服务
func NewGroupService(repo *GroupRepository) *GroupService {
	return &GroupService{repo: repo}
}

// FindTreeList 获取文件分组树列表
func (s *GroupService) FindTreeList(ctx context.Context, storeId uint32) ([]*UploadGroup, error) {
	filter := &GroupFilter{StoreId: storeId}
	list, err := s.repo.FindAll(ctx, filter, nil, nil)
	if err != nil {
		return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("获取分组列表失败"))
	}
	if len(list) > 0 {
		groupList := GroupList{}
		for _, group := range list {
			groupList = append(groupList, group)
		}
		list = groupList.Tree()
	}
	return list, nil
}

// Create 创建文件分组
func (s *GroupService) Create(ctx context.Context, req *GroupCreateRequest) error {
	if err := s.checkName(ctx, req.Name, req.StoreId); err != nil {
		return err
	}
	group := &UploadGroup{StoreId: req.StoreId, Name: req.Name, ParentId: req.ParentId, Sort: req.Sort}
	if err := s.repo.Create(ctx, group); err != nil {
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("创建分组失败"))
	}
	return nil
}

// checkName 检查文件分组名称
func (s *GroupService) checkName(ctx context.Context, name string, storeId uint32) error {
	filter := &GroupFilter{Name: name, StoreId: storeId}
	_, err := s.repo.FindOne(ctx, filter)
	if err != nil {
		if errors.Is(err, baserepo.ErrRecordNotFound) {
			return nil
		}
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("检查分组名称失败"))
	}
	return apperror.New(errcode.Conflict, apperror.WithMsg("分组名称已存在"))
}

// Update 更新文件分组
func (s *GroupService) Update(ctx context.Context, req *GroupUpdateRequest) error {
	filter := &GroupFilter{Id: req.ID, StoreId: req.StoreId}
	item, err := s.repo.FindOne(ctx, filter)
	if err != nil {
		if errors.Is(err, baserepo.ErrRecordNotFound) {
			return apperror.New(errcode.NotFound, apperror.WithMsg("分组不存在或无权限访问"))
		}
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("更新分组失败"))
	}
	if item.Name != req.Name {
		if err := s.checkName(ctx, req.Name, req.StoreId); err != nil {
			return err
		}
	}
	updateData := map[string]any{"name": req.Name, "parent_id": req.ParentId, "sort": req.Sort, "store_id": req.StoreId}
	if err := s.repo.Updates(ctx, item, updateData); err != nil {
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("更新分组失败"))
	}
	return nil
}

// Delete 删除文件分组
func (s *GroupService) Delete(ctx context.Context, req *GroupDeleteRequest) error {
	filter := &GroupFilter{Id: req.ID, StoreId: req.StoreId}
	_, err := s.repo.FindOne(ctx, filter)
	if err != nil {
		if errors.Is(err, baserepo.ErrRecordNotFound) {
			return apperror.New(errcode.NotFound, apperror.WithMsg("分组不存在或无权限访问"))
		}
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("删除分组失败"))
	}
	if err := s.repo.Delete(ctx, req.ID); err != nil {
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("删除分组失败"))
	}
	return nil
}

// FileService 文件服务
type FileService struct {
	repo    *FileRepository
	settSvc SettingProvider
}

// NewFileService 创建文件服务
func NewFileService(repo *FileRepository, settSvc SettingProvider) *FileService {
	return &FileService{repo: repo, settSvc: settSvc}
}

// FindList 获取文件列表
func (s *FileService) FindList(ctx context.Context, req *FileListRequest) (*ListResult, error) {
	result := &ListResult{List: []*UploadFile{}, Total: 0}
	filter := &FileFilter{StoreId: req.StoreId, GroupId: req.GroupId, FileType: req.FileType, FileName: req.FileName}
	pagination := baserepo.NewPagination(req.Page, req.Limit)
	orders := baserepo.Orders{{Field: "id", Sort: "desc"}}
	total, err := s.repo.Count(ctx, filter)
	if err != nil {
		return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("获取文件列表失败"))
	}
	if total == 0 {
		return result, nil
	}
	result.Total = total
	list, err := s.repo.FindAll(ctx, filter, pagination, orders)
	if err != nil {
		return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("获取文件列表失败"))
	}
	result.List = list
	return result, nil
}

// Detail 获取文件详情
func (s *FileService) Detail(ctx context.Context, req *FileDetailRequest) (*UploadFile, error) {
	filter := &FileFilter{Id: req.ID, StoreId: req.StoreId}
	item, err := s.repo.FindOne(ctx, filter)
	if err != nil {
		if errors.Is(err, baserepo.ErrRecordNotFound) {
			return nil, apperror.New(errcode.NotFound, apperror.WithMsg("文件不存在"))
		}
		return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("获取文件详情失败"))
	}
	return item, nil
}

// getUploadConfig 获取上传配置
func (s *FileService) getUploadConfig(ctx context.Context, storeId uint32) (*UploadConfig, error) {
	config := &UploadConfig{}
	if err := s.settSvc.GetSettingValue(ctx, "upload", storeId, config); err != nil {
		return nil, err
	}
	return config, nil
}

// sanitizeFilename 清理文件名
func (s *FileService) sanitizeFilename(filename string) string {
	filename = filepath.Base(filename)
	filename = strings.ReplaceAll(filename, " ", "_")
	filename = strings.ReplaceAll(filename, "..", "")
	reg := regexp.MustCompile(`[^\w\-\.]`)
	filename = reg.ReplaceAllString(filename, "")
	return filename
}

// validateFileContent 验证文件内容
func (s *FileService) validateFileContent(fileHeader []byte, fileExt string, allowedTypes []string) error {
	detectedType := http.DetectContentType(fileHeader)
	if detectedType == "application/octet-stream" {
		return apperror.New(errcode.InvalidInput, apperror.WithMsg("无法识别的文件类型"))
	}
	if slices.Contains(allowedTypes, "image") {
		if strings.HasPrefix(detectedType, "image/") {
			return nil
		}
	}
	if slices.Contains(allowedTypes, "document") {
		switch {
		case strings.HasPrefix(detectedType, "application/pdf"):
			return nil
		case strings.HasPrefix(detectedType, "application/msword"):
			return nil
		case strings.HasPrefix(detectedType, "application/vnd.ms-excel"):
			return nil
		}
	}
	if slices.Contains(allowedTypes, "video") {
		if strings.HasPrefix(detectedType, "video/") {
			return nil
		}
	}
	if slices.Contains(allowedTypes, "archive") {
		switch fileExt {
		case "zip", "rar":
			return nil
		}
	}
	return apperror.New(errcode.InvalidInput, apperror.WithMsg("文件内容类型与扩展名不匹配"))
}

// generateFilePath 生成文件路径
func (s *FileService) generateFilePath(file *multipart.FileHeader) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, src); err != nil {
		return "", err
	}
	md5Str := hex.EncodeToString(hash.Sum(nil))

	ext := filepath.Ext(file.Filename)
	if ext == "" {
		ext = ".dat"
	}

	path := filepath.Join("uploads", md5Str[:2], md5Str[2:4], md5Str[4:]+ext)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return "", err
	}
	return path, nil
}

// Upload 上传文件
func (s *FileService) Upload(ctx context.Context, req *FileRequest) (*UploadFile, error) {
	config, err := s.getUploadConfig(ctx, req.StoreId)
	if err != nil {
		return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("获取上传配置失败"))
	}

	maxSize, _ := strconv.Atoi(config.MaxSize)
	if req.File.Size > int64(maxSize)*1024*1024 {
		return nil, apperror.New(errcode.InvalidInput, apperror.WithMsg(fmt.Sprintf("文件大小不能超过%dMB", maxSize)))
	}

	fileExt := strings.ToLower(strings.TrimPrefix(filepath.Ext(req.File.Filename), "."))
	if !s.checkFileExt(config.AllowedTypes, fileExt) {
		return nil, apperror.New(errcode.InvalidInput, apperror.WithMsg("不支持的文件类型"))
	}

	fileHeader := make([]byte, 512)
	src, err := req.File.Open()
	if err != nil {
		return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("读取文件失败"))
	}
	if _, err = src.Read(fileHeader); err != nil {
		src.Close()
		return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("读取文件失败"))
	}
	src.Close()

	if err := s.validateFileContent(fileHeader, fileExt, config.AllowedTypes); err != nil {
		return nil, err
	}

	safeFilename := s.sanitizeFilename(req.File.Filename)
	if safeFilename == "" {
		return nil, apperror.New(errcode.InvalidInput, apperror.WithMsg("无效的文件名"))
	}

	savePath, err := s.generateFilePath(req.File)
	if err != nil {
		return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("生成文件路径失败"))
	}

	fileType := int8(20)
	if strings.HasPrefix(http.DetectContentType(fileHeader), "image/") {
		fileType = 10
	} else if strings.HasPrefix(http.DetectContentType(fileHeader), "video/") {
		fileType = 30
	}

	filePath := "/" + strings.ReplaceAll(savePath, "\\", "/")
	uploadFile := &UploadFile{
		GroupId: req.GroupId, Channel: 10, Storage: config.StorageType,
		Domain: "", FileType: fileType, FilePath: filePath,
		FileName: req.File.Filename, FileSize: uint32(req.File.Size),
		FileExt: fileExt, UploaderId: req.UploaderId, StoreId: req.StoreId,
	}

	if config.StorageType == "qiniu" {
		qiniu := &QiniuConfig{}
		if err := s.settSvc.GetSettingValue(ctx, "qiniu", req.StoreId, qiniu); err != nil {
			return nil, err
		}
		if !qiniu.IsEnabled {
			return nil, apperror.New(errcode.Internal, apperror.WithMsg("七牛云存储未启用"))
		}
		ctx = WithQiniuConfig(ctx, qiniu)
	}

	uploader, err := NewUploader(config.StorageType, ctx)
	if err != nil {
		return nil, err
	}

	domain, err := uploader.Upload(ctx, req.File, savePath)
	if err != nil {
		return nil, err
	}
	uploadFile.Domain = domain

	if err := s.repo.Create(ctx, uploadFile); err != nil {
		return nil, apperror.Wrap(errcode.Internal, err, apperror.WithMsg("创建文件记录失败"))
	}

	return uploadFile, nil
}

// checkFileExt 检查文件扩展名
func (s *FileService) checkFileExt(allowedTypes []string, fileExt string) bool {
	for _, allowedType := range allowedTypes {
		switch allowedType {
		case "image":
			if slices.Contains([]string{"jpg", "png", "gif"}, fileExt) {
				return true
			}
		case "document":
			if slices.Contains([]string{"pdf", "doc", "xls", "xlsx"}, fileExt) {
				return true
			}
		case "video":
			if slices.Contains([]string{"mp4", "mov"}, fileExt) {
				return true
			}
		case "archive":
			if slices.Contains([]string{"zip", "rar"}, fileExt) {
				return true
			}
		}
	}
	return false
}

// Delete 删除文件
func (s *FileService) Delete(ctx context.Context, req *FileDeleteRequest) error {
	filter := &FileFilter{Id: req.ID, StoreId: req.StoreId}
	_, err := s.repo.FindOne(ctx, filter)
	if err != nil {
		if errors.Is(err, baserepo.ErrRecordNotFound) {
			return apperror.New(errcode.NotFound, apperror.WithMsg("文件不存在"))
		}
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("删除文件失败"))
	}
	if err := s.repo.Delete(ctx, req.ID); err != nil {
		return apperror.Wrap(errcode.Internal, err, apperror.WithMsg("删除文件失败"))
	}
	return nil
}
