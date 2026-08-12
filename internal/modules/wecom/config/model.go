package config

import "gorm.io/plugin/soft_delete"

// WecomConfig 企业微信配置
type WecomConfig struct {
	ID         uint32 `json:"id" gorm:"primaryKey"`
	StoreId    uint32 `json:"store_id" gorm:"not null;default:0;comment:企业ID ( 租户 ) ;index:idx_store_id"`
	CorpId     string `json:"corp_id" gorm:"size:64;not null;default:'';comment:企业微信企业ID ( CorpID ) "`
	CorpName   string `json:"corp_name" gorm:"size:128;not null;default:'';comment:企业微信企业名称 ( 展示用 ) "`
	ApiBaseUrl string `json:"api_base_url" gorm:"size:255;not null;default:'';comment:企微API代理地址 ( 留空使用官方地址 ) "`
	Status     int8   `json:"status" gorm:"type:tinyint;not null;default:0;comment:接入状态 ( 0未接入 1已接入 2已停用 ) "`
	CreatedAt  uint32 `json:"created_at" gorm:"not null;comment:创建时间;autoCreateTime"`
	UpdatedAt  uint32 `json:"updated_at" gorm:"not null;comment:更新时间;autoUpdateTime"`

	DeletedAt soft_delete.DeletedAt `json:"-" gorm:"not null;default:0;comment:删除时间"`
}

// WecomApp 企业微信应用
type WecomApp struct {
	ID             uint32 `json:"id" gorm:"primaryKey"`
	StoreId        uint32 `json:"store_id" gorm:"not null;default:0;comment:企业ID ( 租户 ) ;index:idx_store_app,priority:1"`
	AppType        int8   `json:"app_type" gorm:"type:tinyint;not null;default:10;comment:类型 ( 10自建应用 20客户联系 30通讯录同步 ) ;index:idx_store_app,priority:2"`
	AppName        string `json:"app_name" gorm:"size:64;not null;default:'';comment:应用名称"`
	CallbackToken  string `json:"callback_token" gorm:"size:32;not null;default:'';comment:回调URL路由标识 ( 系统随机生成, 用于定位应用 ) ;index:idx_callback_token"`
	AgentId        uint32 `json:"agent_id" gorm:"not null;default:0;comment:应用ID ( AgentID, 客户联系/通讯录同步为0 ) ;index:idx_store_app,priority:3"`
	Secret         string `json:"secret" gorm:"size:255;not null;default:'';comment:应用密钥 ( 建议加密存储 ) "`
	Token          string `json:"token" gorm:"size:64;not null;default:'';comment:企微回调Token ( 用户配置, 用于验签解密 ) "`
	EncodingAesKey string `json:"encoding_aes_key" gorm:"size:64;not null;default:'';comment:回调EncodingAESKey ( 建议加密存储 ) "`
	Status         int8   `json:"status" gorm:"type:tinyint;not null;default:1;comment:状态 ( 1启用 0停用 ) "`
	CreatedAt      uint32 `json:"created_at" gorm:"not null;comment:创建时间;autoCreateTime"`
	UpdatedAt      uint32 `json:"updated_at" gorm:"not null;comment:更新时间;autoUpdateTime"`

	DeletedAt soft_delete.DeletedAt `json:"-" gorm:"not null;default:0;comment:删除时间"`
}
