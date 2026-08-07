package rbac

// ListResult 列表数据结构体
type ListResult struct {
	List  any   `json:"list"`
	Total int64 `json:"total"`
}

// RbacMenuCreateRequest 创建菜单请求参数
type RbacMenuCreateRequest struct {
	Name       string `json:"name" validate:"required,min=2,max=20"`
	Type       int8   `json:"type" validate:"required,gt=0"`
	ParentId   uint32 `json:"parent_id"`
	Path       string `json:"path" validate:"required"`
	IsPage     int8   `json:"is_page"`
	ModuleKey  string `json:"module_key"`
	ActionMark string `json:"action_mark"`
	Sort       uint32 `json:"sort"`
}

// RbacMenuUpdateRequest 编辑菜单请求参数
type RbacMenuUpdateRequest struct {
	ID         uint32 `json:"id" validate:"required"`
	Name       string `json:"name" validate:"required,min=2,max=20"`
	Type       int8   `json:"type" validate:"required,gt=0"`
	ParentId   uint32 `json:"parent_id"`
	Path       string `json:"path" validate:"required"`
	IsPage     int8   `json:"is_page"`
	ModuleKey  string `json:"module_key"`
	ActionMark string `json:"action_mark"`
	Sort       uint32 `json:"sort"`
}

// RbacMenuSyncRequest 同步菜单请求参数
type RbacMenuSyncRequest struct {
	Name      string `json:"name" validate:"required,min=2,max=20"`
	Path      string `json:"path" validate:"required"`
	ModuleKey string `json:"module_key" validate:"required"`
	Sort      uint32 `json:"sort"`

	Children []RbacMenuSyncRequest       `json:"children"`
	Actions  []RbacMenuActionSyncRequest `json:"actions"`
}

// RbacMenuActionSyncRequest 同步操作菜单请求参数
type RbacMenuActionSyncRequest struct {
	Name       string `json:"name" validate:"required,min=2,max=20"`
	ActionMark string `json:"action_mark" validate:"required"`
}

// RbacMenuDeleteRequest 删除数据
type RbacMenuDeleteRequest struct {
	ID uint32 `json:"id" validate:"required"`
}

// RbacMenuApiListRequest 获取菜单关联的api列表
type RbacMenuApiListRequest struct {
	MenuID uint32 `json:"menu_id" validate:"required"`
}

// RbacMenuApiSaveRequest 保存菜单关联的api列表
type RbacMenuApiSaveRequest struct {
	MenuID uint32   `json:"menu_id" validate:"required"`
	ApiIds []uint32 `json:"api_ids"`
}

// RbacApiCreateRequest 创建接口权限
type RbacApiCreateRequest struct {
	Name     string `json:"name" validate:"required,min=2,max=20"`
	Url      string `json:"url" validate:"required"`
	ParentId uint32 `json:"parent_id"`
	Sort     uint32 `json:"sort"`
}

// RbacApiUpdateRequest 创建接口权限
type RbacApiUpdateRequest struct {
	ID       uint32 `json:"id" validate:"required"`
	Name     string `json:"name" validate:"required,min=2,max=20"`
	Url      string `json:"url" validate:"required"`
	ParentId uint32 `json:"parent_id"`
	Sort     uint32 `json:"sort"`
}

// RbacApiDeleteRequest 删除数据
type RbacApiDeleteRequest struct {
	ID uint32 `json:"id" validate:"required"`
}

// RbacStoreListRequest 获取企业列表请求参数
type RbacStoreListRequest struct {
	Page      int    `json:"page" validate:"required,min=1"`
	Limit     int    `json:"limit" validate:"required,min=1,max=1000"`
	Name      string `json:"name"`
	IsRecycle int8   `json:"is_recycle" validate:"oneof=0 1 -1"`
}

// RbacStoreCreateRequest 创建企业请求参数
type RbacStoreCreateRequest struct {
	Name         string `json:"name" validate:"required,min=2,max=50"`
	ShortName    string `json:"short_name" validate:"required,min=2,max=50"`
	Contact      string `json:"contact" validate:"required,min=2,max=50"`
	ContactPhone string `json:"contact_phone" validate:"required,min=11,max=20"`
	Description  string `json:"description" validate:"max=500"`
	LogoImageID  uint32 `json:"logo_image_id"`
	Sort         uint32 `json:"sort"`
	Username     string `json:"username" validate:"required,min=5,max=20,alphanum"`
}

// RbacStoreCreateResponse 创建企业响应
type RbacStoreCreateResponse struct {
	StoreId  uint32 `json:"store_id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// RbacStoreUpdateRequest 编辑企业请求参数
type RbacStoreUpdateRequest struct {
	ID           uint32 `json:"id" validate:"required"`
	Name         string `json:"name" validate:"required,min=2,max=50"`
	ShortName    string `json:"short_name" validate:"required,min=2,max=50"`
	Contact      string `json:"contact" validate:"required,min=2,max=50"`
	ContactPhone string `json:"contact_phone" validate:"required,min=11,max=20"`
	Description  string `json:"description" validate:"max=500"`
	LogoImageID  uint32 `json:"logo_image_id"`
	Sort         uint32 `json:"sort"`
}

// RbacStoreDeleteRequest 删除企业请求参数
type RbacStoreDeleteRequest struct {
	ID uint32 `json:"id" validate:"required"`
}

// RbacRoleListRequest 获取角色列表请求参数
type RbacRoleListRequest struct {
	RoleName string `json:"role_name"`
	StoreId  uint32 `json:"store_id"`
	IsSuper  int8   `json:"-"`
}

// RbacRolePageListRequest 平台端角色分页列表请求参数
type RbacRolePageListRequest struct {
	RoleName string `json:"role_name"`
	StoreId  uint32 `json:"store_id"`
	Page     int    `json:"page" validate:"required,min=1"`
	Limit    int    `json:"limit" validate:"required,min=1,max=100"`
	IsSuper  int8   `json:"-"`
}

// RbacRoleCreateRequest 创建角色请求参数
type RbacRoleCreateRequest struct {
	RoleName string `json:"role_name" validate:"required,min=2,max=50"`
	ParentId uint32 `json:"parent_id"`
	Sort     uint32 `json:"sort"`
	StoreId  uint32 `json:"store_id"`
	IsSuper  int8   `json:"-"`
}

// RbacRoleUpdateRequest 更新角色请求参数
type RbacRoleUpdateRequest struct {
	ID       uint32 `json:"id" validate:"required"`
	RoleName string `json:"role_name" validate:"required,min=2,max=50"`
	ParentId uint32 `json:"parent_id"`
	Sort     uint32 `json:"sort"`
	StoreId  uint32 `json:"store_id"`
	IsSuper  int8   `json:"-"`
}

// RbacRoleDeleteRequest 删除角色请求参数
type RbacRoleDeleteRequest struct {
	ID      uint32 `json:"id" validate:"required"`
	StoreId uint32 `json:"store_id"`
}

// RbacUserRoleSetRequest 设置用户角色请求参数
type RbacUserRoleSetRequest struct {
	UserID  uint32   `json:"user_id" validate:"required"`
	RoleIDs []uint32 `json:"role_ids" validate:"required"`
	StoreId uint32   `json:"store_id"`
}

// RbacRoleMenuSetRequest 设置角色菜单请求参数
type RbacRoleMenuSetRequest struct {
	RoleID  uint32   `json:"role_id" validate:"required"`
	MenuIDs []uint32 `json:"menu_ids" validate:"required"`
	StoreId uint32   `json:"store_id"`
}

// RbacUserListRequest 获取用户列表请求参数
type RbacUserListRequest struct {
	StoreId  uint32 `json:"store_id"`
	Username string `json:"username"`
	RealName string `json:"real_name"`
	Page     int    `json:"page" validate:"required,min=1"`
	Limit    int    `json:"limit" validate:"required,min=1,max=100"`
	IsSuper  int8   `json:"-"`
}

// RbacUserCreateRequest 创建用户请求参数
type RbacUserCreateRequest struct {
	Username string `json:"username" validate:"required,min=5,max=20,alphanum"`
	Password string `json:"password" validate:"required,min=6,max=20"`
	RealName string `json:"real_name" validate:"required,min=2,max=20"`
	IsSuper  int8   `json:"-"`
	StoreId  uint32 `json:"store_id"`
	Sort     uint32 `json:"sort"`
}

// RbacUserUpdateRequest 更新用户请求参数
type RbacUserUpdateRequest struct {
	ID       uint32 `json:"id" validate:"required"`
	Username string `json:"username" validate:"required,min=5,max=20,alphanum"`
	RealName string `json:"real_name" validate:"required,min=2,max=20"`
	IsSuper  int8   `json:"-"`
	StoreId  uint32 `json:"store_id" validate:"required"`
	Sort     uint32 `json:"sort"`
}

// RbacUserDeleteRequest 删除用户请求参数
type RbacUserDeleteRequest struct {
	ID      uint32 `json:"id" validate:"required"`
	StoreId uint32 `json:"store_id"`
}

// RbacUserResetPasswordRequest 重置用户密码请求参数
type RbacUserResetPasswordRequest struct {
	ID      uint32 `json:"id" validate:"required"`
	StoreId uint32 `json:"store_id"`
}

// MenuApiRelationResponse 菜单接口权限关系
type MenuApiRelationResponse struct {
	MenuId uint32   `json:"menu_id"`
	ApiIds []uint32 `json:"api_ids"`
}

// AuthLoginRequest 登录参数
type AuthLoginRequest struct {
	Username    string `json:"username" validate:"required"`
	Password    string `json:"password" validate:"required"`
	CaptchaID   string `json:"captcha_id" validate:"required"`
	CaptchaCode string `json:"captcha_code" validate:"required"`
}

// AuthGetPermissionsRequest 获取权限参数
type AuthGetPermissionsRequest struct {
	IsTree bool `json:"is_tree"`
}

// ChangePasswordRequest 修改密码请求参数
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=6,max=20"`
}

// AdminLoginResponse 登录返回
type AdminLoginResponse struct {
	Token string    `json:"token"`
	Ttl   int       `json:"ttl"`
	User  *RbacUser `json:"user,omitempty"`
}

// CaptchaGenerateResponse 验证码生成响应
type CaptchaGenerateResponse struct {
	CaptchaID   string `json:"captcha_id"`
	MasterImage string `json:"master_image"`
	ThumbImage  string `json:"thumb_image"`
}

// CaptchaClickPoint 验证码点选坐标
type CaptchaClickPoint struct {
	X int `json:"x"`
	Y int `json:"y"`
}
