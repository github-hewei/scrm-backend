package contact

// DepartmentNode 部门树节点
type DepartmentNode struct {
	ID           uint32            `json:"id"`
	DepartmentId uint32            `json:"department_id"`
	ParentId     uint32            `json:"parent_id"`
	Name         string            `json:"name"`
	Sort         uint32            `json:"sort"`
	MemberCount  uint32            `json:"member_count"`
	Children     []*DepartmentNode `json:"children"`
}

// DepartmentTreeResponse 部门树响应
type DepartmentTreeResponse struct {
	Tree []*DepartmentNode `json:"tree"`
}

// MemberListRequest 成员列表请求
type MemberListRequest struct {
	Page         int    `json:"page" validate:"required,min=1"`
	Limit        int    `json:"limit" validate:"required,min=1,max=100"`
	DepartmentId uint32 `json:"department_id"`
	Status       int8   `json:"status"`
}

// MemberInfo 成员信息
type MemberInfo struct {
	UserId      string   `json:"user_id"`
	Name        string   `json:"name"`
	Position    string   `json:"position"`
	Avatar      string   `json:"avatar"`
	Status      int8     `json:"status"`
	Departments []string `json:"departments"` // 所属部门名列表
	IsLeader    bool     `json:"is_leader"`   // 是否部门负责人
}

// MemberListResponse 成员列表响应
type MemberListResponse struct {
	List  []*MemberInfo `json:"list"`
	Total int64         `json:"total"`
}

// MemberDetailRequest 成员详情请求
type MemberDetailRequest struct {
	UserId string `json:"user_id" validate:"required,max=64"`
}

// MemberDepartmentInfo 成员所属部门信息
type MemberDepartmentInfo struct {
	DepartmentId uint32 `json:"department_id"`
	Name         string `json:"name"`
	IsLeader     bool   `json:"is_leader"`
}

// MemberDetailResponse 成员详情响应
type MemberDetailResponse struct {
	UserId      string                  `json:"user_id"`
	Name        string                  `json:"name"`
	Position    string                  `json:"position"`
	Mobile      string                  `json:"mobile"`
	Gender      string                  `json:"gender"`
	Email       string                  `json:"email"`
	Avatar      string                  `json:"avatar"`
	Status      int8                    `json:"status"`
	Departments []*MemberDepartmentInfo `json:"departments"`
}
