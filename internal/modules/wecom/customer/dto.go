package customer

// CustomerListRequest 客户列表请求
type CustomerListRequest struct {
	Page  int  `json:"page" validate:"required,min=1"`
	Limit int  `json:"limit" validate:"required,min=1,max=100"`
	Type  int8 `json:"type"`
}

// CustomerInfo 客户列表项
type CustomerInfo struct {
	ExternalUserid string `json:"external_userid"`
	Name           string `json:"name"`
	Avatar         string `json:"avatar"`
	Type           int8   `json:"type"`
	Gender         int8   `json:"gender"`
	GenderText     string `json:"gender_text"`
	CorpName       string `json:"corp_name"`
}

// CustomerListResponse 客户列表响应
type CustomerListResponse struct {
	List  []*CustomerInfo `json:"list"`
	Total int64           `json:"total"`
}

// CustomerDetailRequest 客户详情请求
type CustomerDetailRequest struct {
	ExternalUserid string `json:"external_userid" validate:"required,max=64"`
}

// FollowInfo 跟进人信息
type FollowInfo struct {
	UserId         string `json:"user_id"`
	UserName       string `json:"user_name"`
	Remark         string `json:"remark"`
	Description    string `json:"description"`
	RemarkCorpName string `json:"remark_corp_name"`
	RemarkMobiles  string `json:"remark_mobiles"`
	AddWay         int16  `json:"add_way"`
	AddWayText     string `json:"add_way_text"`
	State          string `json:"state"`
	CreateTime     uint32 `json:"create_time"`
}

// CustomerDetailResponse 客户详情响应
type CustomerDetailResponse struct {
	ExternalUserid string        `json:"external_userid"`
	Name           string        `json:"name"`
	Avatar         string        `json:"avatar"`
	Type           int8          `json:"type"`
	Gender         int8          `json:"gender"`
	GenderText     string        `json:"gender_text"`
	Position       string        `json:"position"`
	CorpName       string        `json:"corp_name"`
	CorpFullName   string        `json:"corp_full_name"`
	Unionid        string        `json:"unionid"`
	Follows        []*FollowInfo `json:"follows"`
}
