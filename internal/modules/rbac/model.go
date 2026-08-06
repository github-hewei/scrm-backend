package rbac

import (
	"gorm.io/plugin/soft_delete"
)

// RbacMenu 菜单表模型
type RbacMenu struct {
	ID         uint32 `json:"id" gorm:"primaryKey"`
	Type       int8   `json:"type" gorm:"type:tinyint;not null;default:10;comment:菜单类型 ( 10页面 20操作 ) "`
	Name       string `json:"name" gorm:"size:255;not null;default:'';comment:菜单名称"`
	Path       string `json:"path" gorm:"size:255;not null;default:'';comment:菜单路径 ( 唯一 ) "`
	IsPage     int8   `json:"is_page" gorm:"type:tinyint;not null;default:1;comment:是否为页面 ( 1是 0否 ) "`
	ModuleKey  string `json:"module_key" gorm:"size:100;not null;default:'';comment:功能模块key"`
	ActionMark string `json:"action_mark" gorm:"size:255;not null;default:'';comment:操作标识"`
	ParentId   uint32 `json:"parent_id" gorm:"not null;default:0;comment:上级菜单ID"`
	Sort       uint32 `json:"sort" gorm:"not null;default:100;comment:排序 ( 数字越小越靠前 ) "`
	CreatedAt  int64  `json:"created_at" gorm:"not null;comment:创建时间;autoCreateTime"`
	UpdatedAt  int64  `json:"updated_at" gorm:"not null;comment:更新时间;autoUpdateTime"`

	Children []*RbacMenu `json:"children" gorm:"-"`
	Actions  []*RbacMenu `json:"actions" gorm:"-"`
}

// RbacMenuList 菜单列表
type RbacMenuList []*RbacMenu

// Tree 构建树形菜单结构
func (list RbacMenuList) Tree() []*RbacMenu {
	idMap := make(map[uint32]*RbacMenu, len(list))
	for _, menu := range list {
		menu.Children = []*RbacMenu{}
		menu.Actions = []*RbacMenu{}
		idMap[menu.ID] = menu
	}

	var roots []*RbacMenu
	for _, menu := range list {
		if menu.ParentId == 0 {
			roots = append(roots, menu)
		} else if parent, ok := idMap[menu.ParentId]; ok {
			switch menu.Type {
			case 10:
				parent.Children = append(parent.Children, menu)
			case 20:
				parent.Actions = append(parent.Actions, menu)
			}
		}
	}

	return roots
}

// RbacApi 接口表模型
type RbacApi struct {
	ID        uint32     `json:"id" gorm:"primaryKey"`
	Name      string     `json:"name" gorm:"size:255;not null;default:'';comment:权限名称"`
	Url       string     `json:"url" gorm:"size:255;not null;default:'';comment:权限url"`
	ParentId  uint32     `json:"parent_id" gorm:"not null;default:0;comment:父级ID"`
	Sort      uint32     `json:"sort" gorm:"not null;default:100;comment:排序 ( 数字越小越靠前 ) "`
	CreatedAt int64      `json:"created_at" gorm:"not null;comment:创建时间;autoCreateTime"`
	UpdatedAt int64      `json:"updated_at" gorm:"not null;comment:更新时间;autoUpdateTime"`
	Children  []*RbacApi `json:"children" gorm:"-"`
}

// RbacApiList 接口列表
type RbacApiList []*RbacApi

// Tree 构建树形接口结构
func (list RbacApiList) Tree() []*RbacApi {
	idMap := make(map[uint32]*RbacApi, len(list))
	for _, menu := range list {
		menu.Children = []*RbacApi{}
		idMap[menu.ID] = menu
	}

	var roots []*RbacApi
	for _, menu := range list {
		if menu.ParentId == 0 {
			roots = append(roots, menu)
		} else if parent, ok := idMap[menu.ParentId]; ok {
			parent.Children = append(parent.Children, menu)
		}
	}

	return roots
}

// RbacMenuApi 菜单权限关联表
type RbacMenuApi struct {
	ID        uint32 `json:"id" gorm:"primaryKey"`
	MenuId    uint32 `json:"menu_id" gorm:"not null;default:0;comment:菜单ID;index:menu_id"`
	ApiId     uint32 `json:"api_id" gorm:"not null;default:0;comment:后台api ID"`
	CreatedAt int64  `json:"created_at" gorm:"not null;comment:创建时间;autoCreateTime"`
}

// RbacRole 角色表模型
type RbacRole struct {
	ID        uint32 `json:"id" gorm:"primaryKey"`
	RoleName  string `json:"role_name" gorm:"size:50;not null;default:'';comment:角色名称"`
	ParentId  uint32 `json:"parent_id" gorm:"not null;default:0;comment:父级角色ID"`
	Sort      uint32 `json:"sort" gorm:"not null;default:100;comment:排序 ( 数字越小越靠前 ) "`
	IsSuper   int8   `json:"is_super" gorm:"type:tinyint;not null;default:0;comment:是否为超级管理员角色"`
	StoreId   uint32 `json:"store_id" gorm:"not null;default:0;comment:企业ID;index:store_id"`
	CreatedAt int64  `json:"created_at" gorm:"not null;comment:创建时间;autoCreateTime"`
	UpdatedAt int64  `json:"updated_at" gorm:"not null;comment:更新时间;autoUpdateTime"`

	Children     []*RbacRole     `json:"children" gorm:"-"`
	RbacRoleMenu []*RbacRoleMenu `json:"rbac_role_menu" gorm:"foreignKey:RoleId"`
}

// RbacRoleList 角色列表
type RbacRoleList []*RbacRole

// Tree 转换为树形结构
func (list RbacRoleList) Tree() []*RbacRole {
	var tree []*RbacRole
	itemMap := make(map[uint32]*RbacRole)

	for _, item := range list {
		itemMap[item.ID] = item
	}

	for _, item := range list {
		if item.ParentId == 0 {
			tree = append(tree, item)
		} else {
			if parent, ok := itemMap[item.ParentId]; ok {
				parent.Children = append(parent.Children, item)
			}
		}
	}

	return tree
}

// RbacRoleMenu 角色关联菜单表模型
type RbacRoleMenu struct {
	ID        uint32 `json:"id" gorm:"primaryKey"`
	RoleId    uint32 `json:"role_id" gorm:"not null;default:0;comment:用户角色ID;index:role_id"`
	MenuId    uint32 `json:"menu_id" gorm:"not null;default:0;comment:菜单ID;index:menu_id"`
	StoreId   uint32 `json:"store_id" gorm:"not null;default:0;comment:企业ID;index:store_id"`
	CreatedAt int64  `json:"created_at" gorm:"not null;comment:创建时间;autoCreateTime"`

	RbacMenu *RbacMenu `json:"rbac_menu" gorm:"foreignKey:MenuId"`
}

// RbacStore 企业表模型
type RbacStore struct {
	ID           uint32 `json:"id" gorm:"primaryKey"`
	Name         string `json:"name" gorm:"size:50;not null;default:'';comment:企业名称"`
	ShortName    string `json:"short_name" gorm:"size:50;not null;default:'';comment:企业简称"`
	Contact      string `json:"contact" gorm:"size:50;not null;default:'';comment:企业联系人"`
	ContactPhone string `json:"contact_phone" gorm:"size:50;not null;default:'';comment:联系电话"`
	Description  string `json:"description" gorm:"size:500;not null;default:'';comment:简介"`
	LogoImageID  uint32 `json:"logo_image_id" gorm:"not null;default:0;comment:LOGO文件ID"`
	Sort         uint32 `json:"sort" gorm:"not null;default:0;comment:排序 ( 数字越小越靠前 ) "`
	IsRecycle    int8   `json:"is_recycle" gorm:"type:tinyint;not null;default:0;comment:是否回收"`
	CreatedAt    int64  `json:"created_at" gorm:"not null;comment:创建时间;autoCreateTime"`
	UpdatedAt    int64  `json:"updated_at" gorm:"not null;comment:更新时间;autoUpdateTime"`

	DeletedAt soft_delete.DeletedAt `json:"-" gorm:"not null;default:0;comment:删除时间"`
}

// RbacUser 管理员表模型
type RbacUser struct {
	ID        uint32 `json:"id" gorm:"primaryKey"`
	Username  string `json:"username" gorm:"size:255;not null;default:'';comment:用户名"`
	Password  string `json:"-" gorm:"size:255;not null;default:'';comment:登录密码"`
	RealName  string `json:"real_name" gorm:"size:255;not null;default:'';comment:姓名"`
	IsSuper   int8   `json:"is_super" gorm:"type:tinyint;not null;default:0;comment:是否为超级管理员"`
	Sort      uint32 `json:"sort" gorm:"not null;default:100;comment:排序 ( 数字越小越靠前 ) "`
	StoreId   uint32 `json:"store_id" gorm:"not null;default:0;comment:企业ID;index:store_id"`
	CreatedAt int64  `json:"created_at" gorm:"not null;comment:创建时间;autoCreateTime"`
	UpdatedAt int64  `json:"updated_at" gorm:"not null;comment:更新时间;autoUpdateTime"`

	DeletedAt    soft_delete.DeletedAt `json:"-" gorm:"not null;default:0;comment:删除时间"`
	RbacUserRole []*RbacUserRole       `json:"rbac_user_role" gorm:"foreignKey:UserId"`
}

// RbacUserRole 用户角色关联表
type RbacUserRole struct {
	ID        uint32 `json:"id" gorm:"primaryKey"`
	UserId    uint32 `json:"user_id" gorm:"not null;default:0;comment:超管用户ID;index:user_id"`
	RoleId    uint32 `json:"role_id" gorm:"not null;default:0;comment:角色ID;index:role_id"`
	StoreId   uint32 `json:"store_id" gorm:"not null;default:0;comment:企业ID;index:store_id"`
	CreatedAt int64  `json:"created_at" gorm:"not null;comment:创建时间;autoCreateTime"`

	RbacRole *RbacRole `json:"rbac_role" gorm:"foreignKey:RoleId"`
}
