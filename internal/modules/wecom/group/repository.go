package group

import (
	"context"

	"github.com/241x/zero-kit/baserepo"
	"gorm.io/gorm"
)

// GroupStatus 客户群状态
type GroupStatus int8

// 群状态定义
const (
	GroupStatusNormal     GroupStatus = 0 // 跟进人正常
	GroupStatusResigned   GroupStatus = 1 // 跟进人离职
	GroupStatusInheriting GroupStatus = 2 // 离职继承中
	GroupStatusInherited  GroupStatus = 3 // 离职继承完成
)

// Name 群状态中文名称
func (s GroupStatus) Name() string {
	switch s {
	case GroupStatusNormal:
		return "正常"
	case GroupStatusResigned:
		return "跟进人离职"
	case GroupStatusInheriting:
		return "离职继承中"
	case GroupStatusInherited:
		return "离职继承完成"
	default:
		return "未知"
	}
}

// GroupMemberType 群成员类型
type GroupMemberType int8

// 群成员类型定义
const (
	GroupMemberTypeInternal GroupMemberType = 1 // 企业成员
	GroupMemberTypeExternal GroupMemberType = 2 // 外部联系人
)

// Name 群成员类型中文名称
func (t GroupMemberType) Name() string {
	switch t {
	case GroupMemberTypeInternal:
		return "企业成员"
	case GroupMemberTypeExternal:
		return "外部联系人"
	default:
		return "未知"
	}
}

// JoinScene 入群方式
type JoinScene int8

// 入群方式定义
const (
	JoinSceneInvite     JoinScene = 1 // 成员直接邀请
	JoinSceneInviteLink JoinScene = 2 // 邀请链接
	JoinSceneScanQr     JoinScene = 3 // 扫描群二维码
)

// Name 入群方式中文名称
func (s JoinScene) Name() string {
	switch s {
	case JoinSceneInvite:
		return "成员直接邀请"
	case JoinSceneInviteLink:
		return "邀请链接"
	case JoinSceneScanQr:
		return "扫描群二维码"
	default:
		return "未知"
	}
}

// WecomGroupRepository 企业微信客户群仓库
type WecomGroupRepository struct {
	*baserepo.BaseRepository[WecomGroup]
}

// NewWecomGroupRepository 创建客户群仓库
func NewWecomGroupRepository(db *gorm.DB) *WecomGroupRepository {
	return &WecomGroupRepository{BaseRepository: baserepo.NewBaseRepository[WecomGroup](db)}
}

// GroupFilter 客户群通用过滤条件（详情/精确查询）
type GroupFilter struct {
	StoreId uint32
	ChatId  string
}

// Apply 应用过滤条件
func (f *GroupFilter) Apply(db *gorm.DB) *gorm.DB {
	if f == nil {
		return db
	}
	if f.StoreId != 0 {
		db = db.Where("store_id = ?", f.StoreId)
	}
	if f.ChatId != "" {
		db = db.Where("chat_id = ?", f.ChatId)
	}
	return db
}

// GroupListFilter 客户群列表过滤条件（status=-1表示全部）
type GroupListFilter struct {
	StoreId uint32
	Owner   string
	Status  int8
}

// Apply 应用过滤条件
func (f *GroupListFilter) Apply(db *gorm.DB) *gorm.DB {
	if f == nil {
		return db
	}
	if f.StoreId != 0 {
		db = db.Where("store_id = ?", f.StoreId)
	}
	if f.Owner != "" {
		db = db.Where("owner = ?", f.Owner)
	}
	if f.Status != -1 {
		db = db.Where("status = ?", f.Status)
	}
	return db
}

// WecomGroupMemberRepository 客户群成员仓库
type WecomGroupMemberRepository struct {
	*baserepo.BaseRepository[WecomGroupMember]
}

// NewWecomGroupMemberRepository 创建客户群成员仓库
func NewWecomGroupMemberRepository(db *gorm.DB) *WecomGroupMemberRepository {
	return &WecomGroupMemberRepository{BaseRepository: baserepo.NewBaseRepository[WecomGroupMember](db)}
}

// FindMemberNames 按成员ID列表查询群内成员名（企业成员），返回 user_id -> name 映射
func (r *WecomGroupMemberRepository) FindMemberNames(ctx context.Context, storeId uint32, userIds []string) (map[string]string, error) {
	result := make(map[string]string, len(userIds))
	if len(userIds) == 0 {
		return result, nil
	}
	list, err := r.FindAll(ctx, &GroupMemberFilter{StoreId: storeId, UserIds: userIds, Type: int8(GroupMemberTypeInternal)}, nil, nil)
	if err != nil {
		return nil, err
	}
	for _, member := range list {
		if member.Name != "" {
			result[member.UserId] = member.Name
		}
	}
	return result, nil
}

// GroupMemberFilter 客户群成员过滤条件
type GroupMemberFilter struct {
	StoreId uint32
	ChatId  string
	UserId  string
	UserIds []string
	Type    int8
}

// Apply 应用过滤条件
func (f *GroupMemberFilter) Apply(db *gorm.DB) *gorm.DB {
	if f == nil {
		return db
	}
	if f.StoreId != 0 {
		db = db.Where("store_id = ?", f.StoreId)
	}
	if f.ChatId != "" {
		db = db.Where("chat_id = ?", f.ChatId)
	}
	if f.UserId != "" {
		db = db.Where("user_id = ?", f.UserId)
	}
	if len(f.UserIds) > 0 {
		db = db.Where("user_id IN ?", f.UserIds)
	}
	if f.Type != 0 {
		db = db.Where("type = ?", f.Type)
	}
	return db
}
