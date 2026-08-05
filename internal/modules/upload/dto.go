package upload

import "mime/multipart"

// ListResult 列表数据结构体
type ListResult struct {
	List  any   `json:"list"`
	Total int64 `json:"total"`
}

// GroupCreateRequest 创建分组请求参数
type GroupCreateRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=50"`
	ParentId uint32 `json:"parent_id"`
	Sort     uint32 `json:"sort"`
	StoreId  uint32 `json:"-"`
}

// GroupUpdateRequest 更新分组请求参数
type GroupUpdateRequest struct {
	ID       uint32 `json:"id" validate:"required"`
	Name     string `json:"name" validate:"required,min=2,max=50"`
	ParentId uint32 `json:"parent_id"`
	Sort     uint32 `json:"sort"`
	StoreId  uint32 `json:"-"`
}

// GroupDeleteRequest 删除分组请求参数
type GroupDeleteRequest struct {
	ID      uint32 `json:"id" validate:"required"`
	StoreId uint32 `json:"store_id"`
}

// FileListRequest 获取文件列表请求参数
type FileListRequest struct {
	GroupId  string `json:"group_id"`
	FileType int8   `json:"file_type"`
	FileName string `json:"file_name"`
	StoreId  uint32 `json:"store_id"`
	Page     int    `json:"page" validate:"required,min=1"`
	Limit    int    `json:"limit" validate:"required,min=1,max=100"`
}

// FileDeleteRequest 删除文件请求参数
type FileDeleteRequest struct {
	ID      uint32 `json:"id" validate:"required"`
	StoreId uint32 `json:"store_id"`
}

// FileRequest 文件上传请求参数
type FileRequest struct {
	File       *multipart.FileHeader `json:"-"`
	GroupId    uint32                `json:"group_id"`
	StoreId    uint32                `json:"store_id"`
	UploaderId uint32                `json:"uploader_id"`
}

// UploadConfig 上传配置
type UploadConfig struct {
	MaxSize      string   `json:"max_size"`
	AllowedTypes []string `json:"allowed_types"`
	StorageType  string   `json:"storage_type"`
}

// QiniuConfig 七牛云配置
type QiniuConfig struct {
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
	Bucket    string `json:"bucket"`
	Domain    string `json:"domain"`
	Zone      string `json:"zone"`
	Expires   uint   `json:"expires"`
	IsEnabled bool   `json:"is_enabled"`
}
