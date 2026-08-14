package sync

import "fmt"

// Scope 同步范围
type Scope string

// 同步范围定义
const (
	ScopeAll     Scope = "all"     // 全部（通讯录+客户+客户群）
	ScopeDept    Scope = "dept"    // 通讯录（部门+成员）
	ScopeContact Scope = "contact" // 外部联系人（客户）
	ScopeGroup   Scope = "group"   // 客户群
)

// ParseScope 解析同步范围参数，返回是否合法
func ParseScope(s string) (Scope, error) {
	switch Scope(s) {
	case ScopeAll, ScopeDept, ScopeContact, ScopeGroup:
		return Scope(s), nil
	default:
		return "", fmt.Errorf("无效的同步范围: %s (可选 all|dept|contact|group)", s)
	}
}
