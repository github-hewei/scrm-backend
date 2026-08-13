package group

// GroupListRequest 客户群列表请求
type GroupListRequest struct {
	Page   int    `json:"page" validate:"required,min=1"`
	Limit  int    `json:"limit" validate:"required,min=1,max=100"`
	Status int8   `json:"status"` // -1全部 0正常 1跟进人离职 2离职继承中 3离职继承完成
	Owner  string `json:"owner"`  // 群主筛选
}

// GroupInfo 客户群列表项
type GroupInfo struct {
	ChatId      string `json:"chat_id"`
	Name        string `json:"name"`
	Owner       string `json:"owner"`
	OwnerName   string `json:"owner_name"`
	MemberCount uint32 `json:"member_count"`
	Status      int8   `json:"status"`
	StatusText  string `json:"status_text"`
	CreateTime  uint32 `json:"create_time"`
}

// GroupListResponse 客户群列表响应
type GroupListResponse struct {
	List  []*GroupInfo `json:"list"`
	Total int64        `json:"total"`
}

// GroupDetailRequest 客户群详情请求
type GroupDetailRequest struct {
	ChatId string `json:"chat_id" validate:"required,max=64"`
}

// GroupMemberInfo 群成员信息
type GroupMemberInfo struct {
	UserId        string `json:"user_id"`
	Name          string `json:"name"`
	Type          int8   `json:"type"`
	TypeText      string `json:"type_text"`
	JoinTime      uint32 `json:"join_time"`
	JoinScene     int8   `json:"join_scene"`
	JoinSceneText string `json:"join_scene_text"`
	IsAdmin       bool   `json:"is_admin"`
}

// GroupDetailResponse 客户群详情响应
type GroupDetailResponse struct {
	ChatId      string             `json:"chat_id"`
	Name        string             `json:"name"`
	Owner       string             `json:"owner"`
	OwnerName   string             `json:"owner_name"`
	Notice      string             `json:"notice"`
	Status      int8               `json:"status"`
	StatusText  string             `json:"status_text"`
	MemberCount uint32             `json:"member_count"`
	CreateTime  uint32             `json:"create_time"`
	Members     []*GroupMemberInfo `json:"members"`
}
