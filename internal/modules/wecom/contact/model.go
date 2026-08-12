package contact

import "gorm.io/plugin/soft_delete"

// WecomDepartment 企业微信部门
type WecomDepartment struct {
	ID           uint32 `json:"id" gorm:"primaryKey"`
	StoreId      uint32 `json:"store_id" gorm:"not null;default:0;comment:企业ID ( 租户 ) ;index:idx_store_department,priority:1;index:idx_store_parent,priority:1"`
	DepartmentId uint32 `json:"department_id" gorm:"not null;default:0;comment:企微部门ID;index:idx_store_department,priority:2"`
	ParentId     uint32 `json:"parent_id" gorm:"not null;default:0;comment:父级部门ID ( 0为根部门 ) ;index:idx_store_parent,priority:2"`
	Name         string `json:"name" gorm:"size:64;not null;default:'';comment:部门名称"`
	Sort         uint32 `json:"sort" gorm:"not null;default:0;comment:部门排序 ( 数字越小越靠前 ) "`
	CreatedAt    uint32 `json:"created_at" gorm:"not null;comment:创建时间;autoCreateTime"`
	UpdatedAt    uint32 `json:"updated_at" gorm:"not null;comment:更新时间;autoUpdateTime"`

	DeletedAt soft_delete.DeletedAt `json:"-" gorm:"not null;default:0;comment:删除时间"`
}

// WecomMember 企业微信成员
type WecomMember struct {
	ID        uint32 `json:"id" gorm:"primaryKey"`
	StoreId   uint32 `json:"store_id" gorm:"not null;default:0;comment:企业ID ( 租户 ) ;index:idx_store_user,priority:1"`
	UserId    string `json:"user_id" gorm:"size:64;not null;default:'';comment:企微成员UserID;index:idx_store_user,priority:2"`
	Name      string `json:"name" gorm:"size:64;not null;default:'';comment:成员姓名"`
	Position  string `json:"position" gorm:"size:255;not null;default:'';comment:职位"`
	Mobile    string `json:"mobile" gorm:"size:30;not null;default:'';comment:手机号"`
	Gender    string `json:"gender" gorm:"size:2;not null;default:'';comment:性别 ( 0未知 1男 2女 ) "`
	Email     string `json:"email" gorm:"size:64;not null;default:'';comment:邮箱"`
	Avatar    string `json:"avatar" gorm:"size:255;not null;default:'';comment:头像URL"`
	Status    int8   `json:"status" gorm:"type:tinyint;not null;default:1;comment:成员状态 ( 1已激活 2已禁用 4未激活 5退出企业 ) "`
	CreatedAt uint32 `json:"created_at" gorm:"not null;comment:创建时间;autoCreateTime"`
	UpdatedAt uint32 `json:"updated_at" gorm:"not null;comment:更新时间;autoUpdateTime"`

	DeletedAt soft_delete.DeletedAt `json:"-" gorm:"not null;default:0;comment:删除时间"`
}

// WecomMemberDepartment 企业微信成员部门
type WecomMemberDepartment struct {
	ID           uint32 `json:"id" gorm:"primaryKey"`
	StoreId      uint32 `json:"store_id" gorm:"not null;default:0;comment:企业ID ( 租户 ) ;index:idx_member,priority:1;index:idx_department,priority:1"`
	UserId       string `json:"user_id" gorm:"size:64;not null;default:'';comment:企微成员UserID ( 关联gaz_wecom_member.user_id ) ;index:idx_member,priority:2"`
	DepartmentId uint32 `json:"department_id" gorm:"not null;default:0;comment:部门ID ( 关联gaz_wecom_department.department_id ) ;index:idx_department,priority:2"`
	IsLeader     int8   `json:"is_leader" gorm:"type:tinyint;not null;default:0;comment:是否该部门负责人 ( 1是 0否 ) "`
	Sort         uint32 `json:"sort" gorm:"not null;default:0;comment:成员在部门内的排序"`
	CreatedAt    uint32 `json:"created_at" gorm:"not null;comment:创建时间;autoCreateTime"`
}
