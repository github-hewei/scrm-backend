package customer

import "gorm.io/plugin/soft_delete"

// WecomCustomer 企业微信客户
type WecomCustomer struct {
	ID             uint32 `json:"id" gorm:"primaryKey"`
	StoreId        uint32 `json:"store_id" gorm:"not null;default:0;comment:企业ID ( 租户 ) ;index:idx_store_customer,priority:1;index:idx_store_unionid,priority:1"`
	ExternalUserid string `json:"external_userid" gorm:"size:64;not null;default:'';comment:企微外部联系人UserID;index:idx_store_customer,priority:2"`
	Name           string `json:"name" gorm:"size:64;not null;default:'';comment:客户名称 ( 昵称 ) "`
	Position       string `json:"position" gorm:"size:128;not null;default:'';comment:职位"`
	Avatar         string `json:"avatar" gorm:"size:255;not null;default:'';comment:头像URL"`
	CorpName       string `json:"corp_name" gorm:"size:128;not null;default:'';comment:公司简称"`
	CorpFullName   string `json:"corp_full_name" gorm:"size:255;not null;default:'';comment:公司全称"`
	Type           int8   `json:"type" gorm:"type:tinyint;not null;default:1;comment:类型 ( 1微信用户 2企业微信用户 ) "`
	Gender         int8   `json:"gender" gorm:"type:tinyint;not null;default:0;comment:性别 ( 0未知 1男 2女 ) "`
	Unionid        string `json:"unionid" gorm:"size:64;not null;default:'';comment:微信UnionID;index:idx_store_unionid,priority:2"`
	CreatedAt      uint32 `json:"created_at" gorm:"not null;comment:创建时间;autoCreateTime"`
	UpdatedAt      uint32 `json:"updated_at" gorm:"not null;comment:更新时间;autoUpdateTime"`

	DeletedAt soft_delete.DeletedAt `json:"-" gorm:"not null;default:0;comment:删除时间"`
}

// WecomCustomerFollow 企业微信客户跟进
type WecomCustomerFollow struct {
	ID             uint32 `json:"id" gorm:"primaryKey"`
	StoreId        uint32 `json:"store_id" gorm:"not null;default:0;comment:企业ID ( 租户 ) ;index:idx_store_follow,priority:1;index:idx_store_user,priority:1;index:idx_store_state,priority:1"`
	ExternalUserid string `json:"external_userid" gorm:"size:64;not null;default:'';comment:客户UserID ( 关联gaz_wecom_customer ) ;index:idx_store_follow,priority:2"`
	UserId         string `json:"user_id" gorm:"size:64;not null;default:'';comment:跟进人企微UserID;index:idx_store_follow,priority:3;index:idx_store_user,priority:2"`
	Remark         string `json:"remark" gorm:"size:128;not null;default:'';comment:跟进人备注"`
	Description    string `json:"description" gorm:"size:255;not null;default:'';comment:跟进人描述"`
	RemarkCorpName string `json:"remark_corp_name" gorm:"size:128;not null;default:'';comment:备注公司名"`
	RemarkMobiles  string `json:"remark_mobiles" gorm:"size:255;not null;default:'';comment:备注手机号 ( 逗号分隔 ) "`
	AddWay         int8   `json:"add_way" gorm:"type:tinyint;not null;default:0;comment:添加方式 ( 与企微add_way一致 ) "`
	State          string `json:"state" gorm:"size:30;not null;default:'';comment:添加渠道标识;index:idx_store_state,priority:2"`
	CreateTime     uint32 `json:"create_time" gorm:"not null;default:0;comment:添加客户时间"`
	CreatedAt      uint32 `json:"created_at" gorm:"not null;comment:创建时间;autoCreateTime"`
	UpdatedAt      uint32 `json:"updated_at" gorm:"not null;comment:更新时间;autoUpdateTime"`

	DeletedAt soft_delete.DeletedAt `json:"-" gorm:"not null;default:0;comment:删除时间"`
}
