package customer

import (
	"github.com/241x/zero-kit/baserepo"
	"gorm.io/gorm"
)

// AddWay 客户添加方式（与企微add_way枚举一致）
type AddWay int16

// 添加方式定义
const (
	AddWayScanQr        AddWay = 1   // 扫描二维码
	AddWaySearchMobile  AddWay = 2   // 搜索手机号
	AddWayCardShare     AddWay = 3   // 名片分享
	AddWayGroupChat     AddWay = 4   // 群聊
	AddWayPhoneContact  AddWay = 5   // 手机通讯录
	AddWayWechatContact AddWay = 6   // 微信联系人
	AddWayThirdApp      AddWay = 8   // 安装第三方应用自动添加
	AddWaySearchEmail   AddWay = 9   // 搜索邮箱
	AddWayVideoChannel  AddWay = 10  // 视频号添加
	AddWaySchedule      AddWay = 11  // 日程参与人
	AddWayMeeting       AddWay = 12  // 会议参与人
	AddWayWechatToWecom AddWay = 13  // 添加微信好友对应的企业微信
	AddWaySmartHardware AddWay = 14  // 智慧硬件专属客服
	AddWayHomeService   AddWay = 15  // 上门服务客服
	AddWayCustomerLink  AddWay = 16  // 获客链接
	AddWayCustomDevelop AddWay = 17  // 定制开发
	AddWayNeedReply     AddWay = 18  // 需求回复
	AddWayThirdPresale  AddWay = 21  // 第三方售前客服
	AddWayBizPartner    AddWay = 22  // 可能的商务伙伴
	AddWayWechatRequest AddWay = 24  // 接受微信好友申请
	AddWayInnerShare    AddWay = 201 // 内部成员共享
	AddWayAssign        AddWay = 202 // 管理员/负责人分配
)

// Name 添加方式中文名称
func (w AddWay) Name() string {
	switch w {
	case AddWayScanQr:
		return "扫描二维码"
	case AddWaySearchMobile:
		return "搜索手机号"
	case AddWayCardShare:
		return "名片分享"
	case AddWayGroupChat:
		return "群聊"
	case AddWayPhoneContact:
		return "手机通讯录"
	case AddWayWechatContact:
		return "微信联系人"
	case AddWayThirdApp:
		return "安装第三方应用自动添加"
	case AddWaySearchEmail:
		return "搜索邮箱"
	case AddWayVideoChannel:
		return "视频号添加"
	case AddWaySchedule:
		return "日程参与人"
	case AddWayMeeting:
		return "会议参与人"
	case AddWayWechatToWecom:
		return "添加微信好友对应的企业微信"
	case AddWaySmartHardware:
		return "智慧硬件专属客服"
	case AddWayHomeService:
		return "上门服务客服"
	case AddWayCustomerLink:
		return "获客链接"
	case AddWayCustomDevelop:
		return "定制开发"
	case AddWayNeedReply:
		return "需求回复"
	case AddWayThirdPresale:
		return "第三方售前客服"
	case AddWayBizPartner:
		return "可能的商务伙伴"
	case AddWayWechatRequest:
		return "接受微信好友申请"
	case AddWayInnerShare:
		return "内部成员共享"
	case AddWayAssign:
		return "管理员/负责人分配"
	default:
		return "未知来源"
	}
}

// Gender 客户性别
type Gender int8

// 性别定义
const (
	GenderUnknown Gender = 0 // 未知
	GenderMale    Gender = 1 // 男
	GenderFemale  Gender = 2 // 女
)

// Name 性别中文名称
func (g Gender) Name() string {
	switch g {
	case GenderMale:
		return "男"
	case GenderFemale:
		return "女"
	default:
		return "未知"
	}
}

// WecomCustomerRepository 企业微信客户仓库
type WecomCustomerRepository struct {
	*baserepo.BaseRepository[WecomCustomer]
}

// NewWecomCustomerRepository 创建客户仓库
func NewWecomCustomerRepository(db *gorm.DB) *WecomCustomerRepository {
	return &WecomCustomerRepository{BaseRepository: baserepo.NewBaseRepository[WecomCustomer](db)}
}

// CustomerFilter 客户过滤条件
type CustomerFilter struct {
	StoreId        uint32
	ExternalUserid string
	Unionid        string
	Type           int8
}

// Apply 应用过滤条件
func (f *CustomerFilter) Apply(db *gorm.DB) *gorm.DB {
	if f == nil {
		return db
	}
	if f.StoreId != 0 {
		db = db.Where("store_id = ?", f.StoreId)
	}
	if f.ExternalUserid != "" {
		db = db.Where("external_userid = ?", f.ExternalUserid)
	}
	if f.Unionid != "" {
		db = db.Where("unionid = ?", f.Unionid)
	}
	if f.Type != 0 {
		db = db.Where("type = ?", f.Type)
	}
	return db
}

// WecomCustomerFollowRepository 客户跟进仓库
type WecomCustomerFollowRepository struct {
	*baserepo.BaseRepository[WecomCustomerFollow]
}

// NewWecomCustomerFollowRepository 创建客户跟进仓库
func NewWecomCustomerFollowRepository(db *gorm.DB) *WecomCustomerFollowRepository {
	return &WecomCustomerFollowRepository{BaseRepository: baserepo.NewBaseRepository[WecomCustomerFollow](db)}
}

// CustomerFollowFilter 客户跟进过滤条件
type CustomerFollowFilter struct {
	StoreId        uint32
	ExternalUserid string
	UserId         string
	State          string
}

// Apply 应用过滤条件
func (f *CustomerFollowFilter) Apply(db *gorm.DB) *gorm.DB {
	if f == nil {
		return db
	}
	if f.StoreId != 0 {
		db = db.Where("store_id = ?", f.StoreId)
	}
	if f.ExternalUserid != "" {
		db = db.Where("external_userid = ?", f.ExternalUserid)
	}
	if f.UserId != "" {
		db = db.Where("user_id = ?", f.UserId)
	}
	if f.State != "" {
		db = db.Where("state = ?", f.State)
	}
	return db
}
