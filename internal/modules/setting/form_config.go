package setting

// FormConfigs 获取系统设置表单配置
func FormConfigs() []FormGroup {
	return []FormGroup{
		{Key: "site", Label: "站点信息", Description: "系统站点基本配置", OnlyPlatform: true, Fields: []FormField{
			{Key: "site_name", Label: "站点名称", Type: "text", Required: true},
			{Key: "domain", Label: "站点域名", Type: "text", Required: true},
		}},
		{Key: "email", Label: "邮件服务", Description: "邮件服务器配置", OnlyPlatform: true, Fields: []FormField{
			{Key: "host", Label: "SMTP服务器", Type: "text", Required: true},
			{Key: "port", Label: "SMTP端口", Type: "text", Required: true},
			{Key: "password", Label: "密码", Type: "text", Required: true},
			{Key: "from_name", Label: "发件人名称", Type: "text"},
			{Key: "from_email", Label: "发件人邮箱", Type: "text"},
		}},
		{Key: "qiniu", Label: "七牛云存储", Description: "七牛云对象存储配置", OnlyPlatform: false, Fields: []FormField{
			{Key: "access_key", Label: "AccessKey", Type: "text", Required: true},
			{Key: "secret_key", Label: "SecretKey", Type: "text", Required: true},
			{Key: "bucket", Label: "存储空间", Type: "text", Required: true},
			{Key: "domain", Label: "访问域名", Type: "text", Required: true},
			{Key: "zone", Label: "存储区域", Type: "select", Options: []FormOption{
				{Label: "华东", Value: "z0"}, {Label: "华北", Value: "z1"}, {Label: "华南", Value: "z2"}, {Label: "北美", Value: "na0"}, {Label: "东南亚", Value: "as0"},
			}},
			{Key: "is_enabled", Label: "是否启用", Type: "switch"},
		}},
		{Key: "upload", Label: "文件上传", Description: "文件上传相关配置", OnlyPlatform: false, Fields: []FormField{
			{Key: "storage_type", Label: "存储方式", Type: "select", Required: true, Options: []FormOption{
				{Label: "本地存储", Value: "local"}, {Label: "七牛云", Value: "qiniu"},
			}},
			{Key: "max_size", Label: "最大文件大小(MB)", Type: "text", Required: true},
			{Key: "allowed_types", Label: "允许的文件类型", Type: "checkbox", Options: []FormOption{
				{Label: "图片(jpg,png,gif)", Value: "image"}, {Label: "文档(pdf,doc,xls)", Value: "document"}, {Label: "视频(mp4,mov)", Value: "video"}, {Label: "压缩包(zip,rar)", Value: "archive"},
			}},
			{Key: "is_enabled", Label: "启用文件上传", Type: "switch"},
		}},
		{Key: "demo", Label: "测试配置", Description: "包含所有表单类型的测试配置", OnlyPlatform: false, Fields: []FormField{
			{Key: "text_input", Label: "文本输入", Type: "text"},
			{Key: "textarea_input", Label: "多行文本", Type: "textarea"},
			{Key: "select_input", Label: "下拉选择", Type: "select", Options: []FormOption{{Label: "选项1", Value: "1"}, {Label: "选项2", Value: "2"}, {Label: "选项3", Value: "3"}}},
			{Key: "checkbox_input", Label: "多选框", Type: "checkbox", Options: []FormOption{{Label: "选项X", Value: "x"}, {Label: "选项Y", Value: "y"}, {Label: "选项Z", Value: "z"}}},
			{Key: "switch_input", Label: "开关", Type: "switch"},
			{Key: "image_input", Label: "图片上传", Type: "image"},
			{Key: "file_input", Label: "文件上传", Type: "file"},
		}},
	}
}
