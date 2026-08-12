package group

import "gorm.io/plugin/soft_delete"

// WecomGroup 企业微信客户群
type WecomGroup struct {
	ID            uint32 `json:"id" gorm:"primaryKey"`
	StoreId       uint32 `json:"store_id" gorm:"not null;default:0;comment:企业ID ( 租户 ) ;index:idx_store_group,priority:1;index:idx_store_owner,priority:1"`
	ChatId        string `json:"chat_id" gorm:"size:64;not null;default:'';comment:客户群ID;index:idx_store_group,priority:2"`
	Name          string `json:"name" gorm:"size:255;not null;default:'';comment:群名"`
	Owner         string `json:"owner" gorm:"size:64;not null;default:'';comment:群主企微UserID;index:idx_store_owner,priority:2"`
	CreateTime    uint32 `json:"create_time" gorm:"not null;default:0;comment:群创建时间"`
	Notice        string `json:"notice" gorm:"size:2000;not null;default:'';comment:群公告"`
	Status        int8   `json:"status" gorm:"type:tinyint;not null;default:0;comment:群状态 ( 0跟进人正常 1跟进人离职 2离职继承中 3离职继承完成 ) "`
	MemberCount   uint32 `json:"member_count" gorm:"not null;default:0;comment:群成员数 ( 冗余, 同步时按member_list长度写入 ) "`
	MemberVersion string `json:"member_version" gorm:"size:64;not null;default:'';comment:成员版本号 ( 详情接口返回, 用于增量同步 ) "`
	CreatedAt     uint32 `json:"created_at" gorm:"not null;comment:创建时间;autoCreateTime"`
	UpdatedAt     uint32 `json:"updated_at" gorm:"not null;comment:更新时间;autoUpdateTime"`

	DeletedAt soft_delete.DeletedAt `json:"-" gorm:"not null;default:0;comment:删除时间"`
}

// WecomGroupMember 企业微信客户群成员
type WecomGroupMember struct {
	ID            uint32 `json:"id" gorm:"primaryKey"`
	StoreId       uint32 `json:"store_id" gorm:"not null;default:0;comment:企业ID ( 租户 ) ;index:idx_store_member,priority:1;index:idx_store_user,priority:1"`
	ChatId        string `json:"chat_id" gorm:"size:64;not null;default:'';comment:客户群ID ( 关联gaz_wecom_group ) ;index:idx_store_member,priority:2"`
	UserId        string `json:"user_id" gorm:"size:64;not null;default:'';comment:成员企微UserID ( 外部联系人为external_userid ) ;index:idx_store_user,priority:2"`
	Type          int8   `json:"type" gorm:"type:tinyint;not null;default:1;comment:成员类型 ( 1企业成员 2外部联系人 ) "`
	Unionid       string `json:"unionid" gorm:"size:64;not null;default:'';comment:微信UnionID ( 外部联系人 ) "`
	JoinTime      uint32 `json:"join_time" gorm:"not null;default:0;comment:入群时间"`
	JoinScene     int8   `json:"join_scene" gorm:"type:tinyint;not null;default:0;comment:入群方式 ( 1成员直接邀请 2邀请链接 3扫描群二维码 ) "`
	InvitorUserid string `json:"invitor_userid" gorm:"size:64;not null;default:'';comment:邀请人企微UserID ( 外部联系人入群时 ) "`
	GroupNickname string `json:"group_nickname" gorm:"size:64;not null;default:'';comment:群内昵称"`
	Name          string `json:"name" gorm:"size:64;not null;default:'';comment:成员名称 ( 外部联系人为微信昵称 ) "`
	IsAdmin       int8   `json:"is_admin" gorm:"type:tinyint;not null;default:0;comment:是否群管理员 ( 1是 0否, 对齐admin_list ) "`
	CreatedAt     uint32 `json:"created_at" gorm:"not null;comment:创建时间;autoCreateTime"`

	DeletedAt soft_delete.DeletedAt `json:"-" gorm:"not null;default:0;comment:删除时间"`
}
