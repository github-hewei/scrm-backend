package config

// ConfigInfo 企业微信配置信息
type ConfigInfo struct {
	CorpId     string `json:"corp_id"`
	CorpName   string `json:"corp_name"`
	ApiBaseUrl string `json:"api_base_url"`
	Status     int8   `json:"status"`
}

// ConfigSaveRequest 保存企业微信配置请求
type ConfigSaveRequest struct {
	CorpId     string `json:"corp_id" validate:"required,max=64"`
	CorpName   string `json:"corp_name" validate:"max=128"`
	ApiBaseUrl string `json:"api_base_url" validate:"max=255"`
}

// AppInfo 应用信息
// CallbackUrl 回显时由站点域名+token拼接返回
type AppInfo struct {
	ID            uint32  `json:"id"`
	AppType       AppType `json:"app_type"`
	AppTypeText   string  `json:"app_type_text"`
	AppName       string  `json:"app_name"`
	AgentId       uint32  `json:"agent_id"`
	SecretSet     bool    `json:"secret_set"`
	TokenSet      bool    `json:"token_set"`
	CallbackToken string  `json:"callback_token"`
	CallbackUrl   string  `json:"callback_url"`
	Status        int8    `json:"status"`
}

// AppSaveRequest 保存应用凭据请求
// CallbackToken 由取地址接口生成后前端传入，首次保存必填
type AppSaveRequest struct {
	AppType        AppType `json:"app_type" validate:"required,oneof=10 20 30"`
	AppName        string  `json:"app_name" validate:"max=64"`
	AgentId        uint32  `json:"agent_id"`
	Secret         string  `json:"secret" validate:"max=255"`
	Token          string  `json:"token" validate:"max=64"`
	EncodingAesKey string  `json:"encoding_aes_key" validate:"max=64"`
	CallbackToken  string  `json:"callback_token" validate:"max=32"`
}

// AppDeleteRequest 删除应用请求
type AppDeleteRequest struct {
	ID uint32 `json:"id" validate:"required"`
}

// AppSaveResponse 保存应用响应
type AppSaveResponse struct {
	ID            uint32 `json:"id"`
	CallbackToken string `json:"callback_token"`
	CallbackUrl   string `json:"callback_url"`
	SecretSet     bool   `json:"secret_set"`
	TokenSet      bool   `json:"token_set"`
}

// CallbackUrlResponse 回调地址响应
// CallbackToken 一次性生成，不入库，保存时由前端回传
type CallbackUrlResponse struct {
	CallbackToken string `json:"callback_token"`
	CallbackUrl   string `json:"callback_url"`
}

// AppListResponse 应用列表响应
type AppListResponse struct {
	List []*AppInfo `json:"list"`
}
