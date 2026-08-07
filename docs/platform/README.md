# 接口文档说明

当前目录存放的是平台端（platform）模块的接口文档，文档遵循 `openapi: 3.1.0` 标准。

## 平台端接口概览

| 分类 | 接口 | 说明 | 所需角色 |
|------|------|------|----------|
| 仪表盘 | `/dashboard/stats` | 平台统计概览 | 超管、运营 |
| 认证管理 | `/captcha/generate` | 生成验证码 | 公开 |
| 认证管理 | `/login` | 平台登录 | 公开 |
| 认证管理 | `/refresh-token` | 刷新令牌 | 已认证 |
| 认证管理 | `/logout` | 退出登录 | 已认证 |
| 认证管理 | `/change-password` | 修改密码 | 已认证 |
| 平台用户管理 | `/platform/user/list` | 平台用户列表 | 超管 |
| 平台用户管理 | `/platform/user/create` | 创建平台用户 | 超管 |
| 平台用户管理 | `/platform/user/update` | 更新平台用户 | 超管 |
| 平台用户管理 | `/platform/user/delete` | 删除平台用户 | 超管 |
| 平台用户管理 | `/platform/user/reset-password` | 重置平台用户密码 | 超管 |
| 设置管理 | `/setting/default/list` | 获取默认设置列表 | 超管、运营 |
| 设置管理 | `/setting/default/create` | 创建默认设置 | 超管、运营 |
| 设置管理 | `/setting/default/update` | 更新默认设置 | 超管、运营 |
| 设置管理 | `/setting/default/delete` | 删除默认设置 | 超管、运营 |
| 设置管理 | `/setting/form-configs` | 获取设置表单配置 | 超管、运营 |
| 上传分组管理 | `/upload/group/list` | 获取分组列表 | 超管、运营 |
| 上传分组管理 | `/upload/group/create` | 创建分组 | 超管、运营 |
| 上传分组管理 | `/upload/group/update` | 更新分组 | 超管、运营 |
| 上传分组管理 | `/upload/group/delete` | 删除分组 | 超管、运营 |
| 上传文件管理 | `/upload/file/list` | 获取文件列表 | 超管、运营 |
| 上传文件管理 | `/upload/file/upload` | 上传文件 | 超管、运营 |
| 上传文件管理 | `/upload/file/detail` | 获取文件详情 | 超管、运营 |
| 上传文件管理 | `/upload/file/delete` | 删除文件 | 超管、运营 |
| 企业管理 | `/rbac/store/list` | 企业列表 | 超管、运营 |
| 企业管理 | `/rbac/store/create` | 创建企业(自动初始化超管) | 超管、运营 |
| 企业管理 | `/rbac/store/update` | 更新企业 | 超管、运营 |
| 企业管理 | `/rbac/store/delete` | 删除企业 | 超管、运营 |
| 企业管理 | `/rbac/store/recycle` | 回收企业 | 超管、运营 |
| 企业管理 | `/rbac/store/restore` | 恢复企业 | 超管、运营 |
| RBAC-菜单管理 | `/rbac/menu/list` | 菜单列表 | 超管、运营 |
| RBAC-菜单管理 | `/rbac/menu/create` | 创建菜单 | 超管、运营 |
| RBAC-菜单管理 | `/rbac/menu/update` | 更新菜单 | 超管、运营 |
| RBAC-菜单管理 | `/rbac/menu/delete` | 删除菜单 | 超管、运营 |
| RBAC-菜单管理 | `/rbac/menu/sync` | 全量同步菜单(页面+操作) | 超管、运营 |
| RBAC-菜单管理 | `/rbac/menu/api/list` | 菜单接口关联列表 | 超管、运营 |
| RBAC-菜单管理 | `/rbac/menu/api/save` | 保存菜单接口关联 | 超管、运营 |
| RBAC-接口管理 | `/rbac/api/list` | 接口列表 | 超管、运营 |
| RBAC-接口管理 | `/rbac/api/create` | 创建接口 | 超管、运营 |
| RBAC-接口管理 | `/rbac/api/update` | 更新接口 | 超管、运营 |
| RBAC-接口管理 | `/rbac/api/delete` | 删除接口 | 超管、运营 |
| RBAC-接口管理 | `/rbac/api/sync` | 全量同步接口 | 超管、运营 |
| RBAC-角色管理 | `/rbac/role/list` | 角色列表(平铺，仅超管角色) | 超管、运营 |
| RBAC-角色管理 | `/rbac/role/create` | 创建角色 | 超管、运营 |
| RBAC-角色管理 | `/rbac/role/update` | 更新角色 | 超管、运营 |
| RBAC-角色管理 | `/rbac/role/delete` | 删除角色 | 超管、运营 |
| RBAC-角色管理 | `/rbac/role/set-menus` | 设置角色菜单 | 超管、运营 |
| 租户用户管理 | `/rbac/user/list` | 租户用户列表 | 超管、运营 |
| 租户用户管理 | `/rbac/user/create` | 创建租户用户 | 超管、运营 |
| 租户用户管理 | `/rbac/user/update` | 更新租户用户 | 超管、运营 |
| 租户用户管理 | `/rbac/user/delete` | 删除租户用户 | 超管、运营 |
| 租户用户管理 | `/rbac/user/set-roles` | 设置租户用户角色 | 超管、运营 |
| 租户用户管理 | `/rbac/user/reset-password` | 重置密码 | 超管、运营 |

### 角色说明

| 角色 | 值 | 权限范围 |
|------|-----|----------|
| 超管 | 0 | 全部权限：平台用户管理 + 租户管理 + 租户用户管理 |
| 运营 | 1 | 租户管理 + 租户用户管理 |
| 审计 | 2 | 暂无接口，预留审计日志查看权限 |
