-- 数据库设计文件：
-- 为了实现数据库设计的版本控制，这个文件需要约定一些规则：
-- 这个文件应该分为多个“执行片段”，用特定的分隔符隔开。
-- 系统启动时依次执行每个“执行片段”的SQL，并记录进度，下次启动时从上次的进度开始执行。
-- 所以“执行片段”在执行之后，不应该再修改。如果需要修改数据库，应该新建一个“执行片段”。
-- 我们将在当前目录创建一个同名文件拼接上 .progress 后缀，用来记录当前进度。
-- 特定的分隔符为：-- [CHECK POINT] --

CREATE TABLE `gaz_rbac_menu` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '菜单ID',
  `type` tinyint(3) unsigned NOT NULL DEFAULT '10' COMMENT '菜单类型(10页面 20操作)',
  `name` varchar(255) NOT NULL DEFAULT '' COMMENT '菜单名称',
  `path` varchar(255) NOT NULL DEFAULT '' COMMENT '菜单路径(唯一)',
  `is_page` tinyint(3) unsigned NOT NULL DEFAULT '1' COMMENT '是否为页面(1是 0否)',
  `module_key` varchar(100) NOT NULL DEFAULT '' COMMENT '功能模块key',
  `action_mark` varchar(255) NOT NULL DEFAULT '' COMMENT '操作标识',
  `parent_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '上级菜单ID',
  `sort` int(11) unsigned NOT NULL DEFAULT '100' COMMENT '排序(数字越小越靠前)',
  `created_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '创建时间',
  `updated_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '更新时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=10000 DEFAULT CHARSET=utf8mb4 COMMENT='菜单记录表';

-- [CHECK POINT] --

CREATE TABLE `gaz_rbac_api` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `name` varchar(255) NOT NULL DEFAULT '' COMMENT '权限名称',
  `url` varchar(255) NOT NULL DEFAULT '' COMMENT '权限url',
  `parent_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '父级ID',
  `sort` int(11) unsigned NOT NULL DEFAULT '100' COMMENT '排序(数字越小越靠前)',
  `created_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '创建时间',
  `updated_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '更新时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=10000 DEFAULT CHARSET=utf8mb4 COMMENT='API权限表';

-- [CHECK POINT] --

CREATE TABLE `gaz_rbac_menu_api` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `menu_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '菜单ID',
  `api_id` int(10) unsigned NOT NULL DEFAULT '0' COMMENT '后台api ID',
  `created_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `menu_id` (`menu_id`)
) ENGINE=InnoDB AUTO_INCREMENT=10000 DEFAULT CHARSET=utf8mb4 COMMENT='菜单与API权限关联表';

-- [CHECK POINT] --

CREATE TABLE `gaz_rbac_user` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `username` varchar(255) NOT NULL DEFAULT '' COMMENT '用户名',
  `password` varchar(255) NOT NULL DEFAULT '' COMMENT '登录密码',
  `real_name` varchar(255) NOT NULL DEFAULT '' COMMENT '姓名',
  `is_super` tinyint(3) unsigned NOT NULL DEFAULT '0' COMMENT '是否为超级管理员',
  `sort` int(11) unsigned NOT NULL DEFAULT '100' COMMENT '排序(数字越小越靠前)',
  `store_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '企业ID',
  `created_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '创建时间',
  `updated_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '更新时间',
  `deleted_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_username` (`username`),
  KEY `store_id` (`store_id`)
) ENGINE=InnoDB AUTO_INCREMENT=10000 DEFAULT CHARSET=utf8mb4 COMMENT='用户记录表';

-- [CHECK POINT] --

CREATE TABLE `gaz_rbac_role` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '角色ID',
  `role_name` varchar(50) NOT NULL DEFAULT '' COMMENT '角色名称',
  `parent_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '父级角色ID',
  `sort` int(11) unsigned NOT NULL DEFAULT '100' COMMENT '排序(数字越小越靠前)',
  `is_super` tinyint(3) unsigned NOT NULL DEFAULT '0' COMMENT '是否为超级管理员角色',
  `store_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '企业ID',
  `created_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '创建时间',
  `updated_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `store_id` (`store_id`)
) ENGINE=InnoDB AUTO_INCREMENT=10000 DEFAULT CHARSET=utf8mb4 COMMENT='用户角色表';

-- [CHECK POINT] --

CREATE TABLE `gaz_rbac_user_role` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `user_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '超管用户ID',
  `role_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '角色ID',
  `store_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '企业ID',
  `created_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `user_id` (`user_id`),
  KEY `role_id` (`role_id`),
  KEY `store_id` (`store_id`)
) ENGINE=InnoDB AUTO_INCREMENT=10000 DEFAULT CHARSET=utf8mb4 COMMENT='用户角色关联记录表';

-- [CHECK POINT] --

CREATE TABLE `gaz_rbac_role_menu` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `role_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '用户角色ID',
  `menu_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '菜单ID',
  `store_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '企业ID',
  `created_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `role_id` (`role_id`),
  KEY `menu_id` (`menu_id`),
  KEY `store_id` (`store_id`)
) ENGINE=InnoDB AUTO_INCREMENT=10000 DEFAULT CHARSET=utf8mb4 COMMENT='角色与菜单权限关系表';

-- [CHECK POINT] --

CREATE TABLE `gaz_rbac_store` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '企业ID',
  `name` varchar(50) NOT NULL DEFAULT '' COMMENT '企业名称',
  `short_name` varchar(50) NOT NULL DEFAULT '' COMMENT '企业简称',
  `contact` varchar(50) NOT NULL DEFAULT '' COMMENT '企业联系人',
  `contact_phone` varchar(50) NOT NULL DEFAULT '' COMMENT '联系电话',
  `description` varchar(500) NOT NULL DEFAULT '' COMMENT '简介',
  `logo_image_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT 'LOGO文件ID',
  `sort` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '排序(数字越小越靠前)',
  `is_recycle` tinyint(3) unsigned NOT NULL DEFAULT '0' COMMENT '是否回收',
  `created_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '创建时间',
  `updated_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '更新时间',
  `deleted_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '删除时间',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=10000 DEFAULT CHARSET=utf8mb4 COMMENT='企业记录表';

-- [CHECK POINT] --

CREATE TABLE `gaz_upload_group` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '分组ID',
  `name` varchar(30) NOT NULL DEFAULT '' COMMENT '分组名称',
  `parent_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '上级分组ID',
  `sort` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '排序(数字越小越靠前)',
  `store_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '企业ID',
  `created_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '创建时间',
  `updated_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '更新时间',
  `deleted_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '删除时间',
  PRIMARY KEY (`id`),
  KEY `store_id` (`store_id`)
) ENGINE=InnoDB AUTO_INCREMENT=10000 DEFAULT CHARSET=utf8mb4 COMMENT='文件库分组表';

-- [CHECK POINT] --

CREATE TABLE `gaz_upload_file` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '文件ID',
  `group_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '文件分组ID',
  `channel` tinyint(3) unsigned NOT NULL DEFAULT '10' COMMENT '上传来源(10后台 20客户端)',
  `storage` varchar(10) NOT NULL DEFAULT '' COMMENT '存储方式',
  `domain` varchar(255) NOT NULL DEFAULT '' COMMENT '存储域名',
  `file_type` tinyint(3) unsigned NOT NULL DEFAULT '10' COMMENT '文件类型(10图片 20附件 30视频)',
  `file_name` varchar(255) NOT NULL DEFAULT '' COMMENT '文件名称(仅显示)',
  `file_path` varchar(255) NOT NULL DEFAULT '' COMMENT '文件路径',
  `file_size` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '文件大小(字节)',
  `file_ext` varchar(20) NOT NULL DEFAULT '' COMMENT '文件扩展名',
  `cover` varchar(255) NOT NULL DEFAULT '' COMMENT '文件封面',
  `uploader_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '上传者用户ID',
  `store_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '企业ID',
  `created_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '创建时间',
  `updated_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '更新时间',
  `deleted_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '删除时间',
  PRIMARY KEY (`id`),
  KEY `group_id` (`group_id`),
  KEY `store_id` (`store_id`)
) ENGINE=InnoDB AUTO_INCREMENT=10000 DEFAULT CHARSET=utf8mb4 COMMENT='文件库表';

-- [CHECK POINT] --

CREATE TABLE `gaz_user` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `username` varchar(32) NOT NULL DEFAULT '' COMMENT '用户名',
  `password` varchar(255) NOT NULL DEFAULT '' COMMENT '密码',
  `mobile` varchar(30) NOT NULL DEFAULT '' COMMENT '用户手机号',
  `nick_name` varchar(64) NOT NULL DEFAULT '' COMMENT '用户昵称',
  `avatar_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '头像文件ID',
  `gender` tinyint(3) unsigned NOT NULL DEFAULT '0' COMMENT '性别',
  `country` varchar(50) NOT NULL DEFAULT '' COMMENT '国家',
  `province` varchar(50) NOT NULL DEFAULT '' COMMENT '省份',
  `city` varchar(50) NOT NULL DEFAULT '' COMMENT '城市',
  `platform` varchar(20) NOT NULL DEFAULT '' COMMENT '注册来源',
  `status` tinyint(3) unsigned NOT NULL DEFAULT '1' COMMENT '账号状态: 1正常 2禁用',
  `last_login_time` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '最后登录时间',
  `store_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '企业ID',
  `created_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '创建时间',
  `updated_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '更新时间',
  `deleted_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '删除时间',
  PRIMARY KEY (`id`),
  KEY `username` (`username`),
  KEY `mobile` (`mobile`),
  KEY `store_id` (`store_id`)
) ENGINE=InnoDB AUTO_INCREMENT=10000 DEFAULT CHARSET=utf8mb4 COMMENT='用户记录表';

-- [CHECK POINT] --

CREATE TABLE `gaz_setting` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `setting_key` varchar(30) NOT NULL COMMENT '设置项标识',
  `setting_values` mediumtext NOT NULL COMMENT '设置内容（json格式）',
  `description` varchar(255) NOT NULL DEFAULT '' COMMENT '设置项描述',
  `store_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '企业ID',
  `created_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '创建时间',
  `updated_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_key` (`store_id`,`setting_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='设置表';

-- [CHECK POINT] --

CREATE TABLE `gaz_setting_default` (
  `id` int(10) unsigned NOT NULL AUTO_INCREMENT,
  `setting_key` varchar(30) NOT NULL COMMENT '设置项标识',
  `setting_values` mediumtext NOT NULL COMMENT '设置内容（json格式）',
  `description` varchar(255) NOT NULL DEFAULT '' COMMENT '设置项描述',
  `created_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '创建时间',
  `updated_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `unique_key` (`setting_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='设置默认值表';

-- [CHECK POINT] --

CREATE TABLE IF NOT EXISTS `gaz_region` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '区划信息ID',
  `name` varchar(255) NOT NULL DEFAULT '' COMMENT '区划名称',
  `pid` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '父级ID',
  `code` varchar(255) NOT NULL DEFAULT '' COMMENT '区划编码',
  `level` tinyint(1) unsigned NOT NULL DEFAULT '1' COMMENT '层级(1省级 2市级 3区/县级)',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='省市区数据表';

-- [CHECK POINT] --

ALTER TABLE `gaz_user`
ADD COLUMN `points` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '用户积分';

-- [CHECK POINT] --

CREATE TABLE `gaz_user_points_log` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `user_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '用户ID',
  `points` int(11) NOT NULL DEFAULT '0' COMMENT '变更积分值',
  `change_type` tinyint(3) unsigned NOT NULL DEFAULT '0' COMMENT '变更类型(1增加 2减少)',
  `source_type` tinyint(3) unsigned NOT NULL DEFAULT '0' COMMENT '来源类型(10消费 20充值 30活动)',
  `source_id` varchar(50) NOT NULL DEFAULT '' COMMENT '来源ID(如订单号)',
  `remark` varchar(255) NOT NULL DEFAULT '' COMMENT '备注',
  `store_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '企业ID',
  `created_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_user_id` (`store_id`, `user_id`),
  KEY `idx_source` (`source_type`,`source_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户积分变更记录表';

-- [CHECK POINT] --

CREATE TABLE `gaz_article_category` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '文章分类ID',
  `name` varchar(50) NOT NULL DEFAULT '' COMMENT '分类名称',
  `status` tinyint(3) NOT NULL DEFAULT '1' COMMENT '状态(1显示 0隐藏)',
  `sort` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '排序方式(数字越小越靠前)',
  `store_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '企业ID',
  `created_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '创建时间',
  `updated_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `store_id` (`store_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='文章分类表';

-- [CHECK POINT] --

CREATE TABLE `gaz_article` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '文章ID',
  `title` varchar(300) NOT NULL DEFAULT '' COMMENT '文章标题',
  `show_type` tinyint(3) unsigned NOT NULL DEFAULT '10' COMMENT '列表显示方式(10小图展示 20大图展示)',
  `category_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '文章分类ID',
  `image_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '封面图ID',
  `content` longtext NOT NULL COMMENT '文章内容',
  `sort` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '文章排序(数字越小越靠前)',
  `status` tinyint(3) unsigned NOT NULL DEFAULT '1' COMMENT '文章状态(0隐藏 1显示)',
  `virtual_views` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '虚拟阅读量(仅用作展示)',
  `actual_views` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '实际阅读量',
  `store_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '企业ID',
  `created_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '创建时间',
  `updated_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '更新时间',
  `deleted_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '是否删除',
  PRIMARY KEY (`id`),
  KEY `category_id` (`category_id`),
  KEY `store_id` (`store_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='文章记录表';

-- [CHECK POINT] --

CREATE TABLE `gaz_platform_user` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `username` varchar(64) NOT NULL DEFAULT '' COMMENT '用户名',
  `password` varchar(255) NOT NULL DEFAULT '' COMMENT '登录密码',
  `real_name` varchar(64) NOT NULL DEFAULT '' COMMENT '姓名',
  `avatar_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '头像文件ID',
  `role` tinyint(3) unsigned NOT NULL DEFAULT '0' COMMENT '角色(0超管 1运营 2审计)',
  `status` tinyint(3) unsigned NOT NULL DEFAULT '1' COMMENT '状态(1启用 0禁用)',
  `last_login_time` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '最后登录时间',
  `last_login_ip` varchar(45) NOT NULL DEFAULT '' COMMENT '最后登录IP',
  `created_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '创建时间',
  `updated_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '更新时间',
  `deleted_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '删除时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `username` (`username`)
) ENGINE=InnoDB AUTO_INCREMENT=10000 DEFAULT CHARSET=utf8mb4 COMMENT='平台管理员表';

-- [CHECK POINT] --

-- 默认密码: 123456
INSERT INTO `gaz_platform_user` (`id`, `username`, `password`, `real_name`, `avatar_id`, `role`, `status`, `last_login_time`, `last_login_ip`, `created_at`, `updated_at`, `deleted_at`) VALUES (10000, 'admin', '$2a$10$VGD3YIumqH75.6Ma6KE58.M2cEtbPQDXG9PN/r2pPU.Pi75jVUoHa', '超级管理员', 0, 0, 1, 0, '', 0, 1786106474, 0);

-- [CHECK POINT] --

-- 角色层级改为扁平管理，移除父级角色ID
ALTER TABLE `gaz_rbac_role` DROP COLUMN `parent_id`;

-- [CHECK POINT] --

-- 企业微信企业配置表：每家企业(租户)一条，仅存储企业身份与回调路由信息
CREATE TABLE `gaz_wecom_config` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `store_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '企业ID(租户)',
  `corp_id` varchar(64) NOT NULL DEFAULT '' COMMENT '企业微信企业ID(CorpID)',
  `corp_name` varchar(128) NOT NULL DEFAULT '' COMMENT '企业微信企业名称(展示用)',
  `api_base_url` varchar(255) NOT NULL DEFAULT '' COMMENT '企微API代理地址(留空使用官方地址)',
  `status` tinyint(3) unsigned NOT NULL DEFAULT '0' COMMENT '接入状态(0未接入 1已接入 2已停用)',
  `created_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '创建时间',
  `updated_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '更新时间',
  `deleted_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_store_id` (`store_id`)
) ENGINE=InnoDB AUTO_INCREMENT=10000 DEFAULT CHARSET=utf8mb4 COMMENT='企业微信企业配置表';

-- [CHECK POINT] --

-- 企业微信应用/功能凭据表：自建应用、客户联系、通讯录同步统一存储，每家企业(租户)可配置多行
CREATE TABLE `gaz_wecom_app` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `store_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '企业ID(租户)',
  `app_type` tinyint(3) unsigned NOT NULL DEFAULT '10' COMMENT '类型(10自建应用 20客户联系 30通讯录同步)',
  `app_name` varchar(64) NOT NULL DEFAULT '' COMMENT '应用名称',
  `callback_token` varchar(32) NOT NULL DEFAULT '' COMMENT '回调URL路由标识(系统随机生成, 用于定位应用)',
  `agent_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '应用ID(AgentID, 客户联系/通讯录同步为0)',
  `secret` varchar(255) NOT NULL DEFAULT '' COMMENT '应用密钥(建议加密存储)',
  `token` varchar(64) NOT NULL DEFAULT '' COMMENT '企微回调Token(用户配置, 用于验签解密)',
  `encoding_aes_key` varchar(64) NOT NULL DEFAULT '' COMMENT '回调EncodingAESKey(建议加密存储)',
  `status` tinyint(3) unsigned NOT NULL DEFAULT '1' COMMENT '状态(1启用 0停用)',
  `created_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '创建时间',
  `updated_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '更新时间',
  `deleted_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_store_app` (`store_id`,`app_type`,`agent_id`),
  KEY `idx_callback_token` (`callback_token`)
) ENGINE=InnoDB AUTO_INCREMENT=10000 DEFAULT CHARSET=utf8mb4 COMMENT='企业微信应用/功能凭据表';

-- [CHECK POINT] --

-- 企业微信部门表：通讯录同步产物，按企微部门ID建树形结构
CREATE TABLE `gaz_wecom_department` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `store_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '企业ID(租户)',
  `department_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '企微部门ID',
  `parent_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '父级部门ID(0为根部门)',
  `name` varchar(64) NOT NULL DEFAULT '' COMMENT '部门名称',
  `sort` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '部门排序(数字越小越靠前)',
  `created_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '创建时间',
  `updated_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '更新时间',
  `deleted_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_store_department` (`store_id`,`department_id`),
  KEY `idx_store_parent` (`store_id`,`parent_id`)
) ENGINE=InnoDB AUTO_INCREMENT=10000 DEFAULT CHARSET=utf8mb4 COMMENT='企业微信部门表';

-- [CHECK POINT] --

-- 企业微信成员表：通讯录同步产物
CREATE TABLE `gaz_wecom_member` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `store_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '企业ID(租户)',
  `user_id` varchar(64) NOT NULL DEFAULT '' COMMENT '企微成员UserID',
  `name` varchar(64) NOT NULL DEFAULT '' COMMENT '成员姓名',
  `position` varchar(255) NOT NULL DEFAULT '' COMMENT '职位',
  `mobile` varchar(30) NOT NULL DEFAULT '' COMMENT '手机号',
  `gender` varchar(2) NOT NULL DEFAULT '' COMMENT '性别(0未知 1男 2女)',
  `email` varchar(64) NOT NULL DEFAULT '' COMMENT '邮箱',
  `avatar` varchar(255) NOT NULL DEFAULT '' COMMENT '头像URL',
  `status` tinyint(3) unsigned NOT NULL DEFAULT '1' COMMENT '成员状态(1已激活 2已禁用 4未激活 5退出企业)',
  `created_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '创建时间',
  `updated_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '更新时间',
  `deleted_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_store_user` (`store_id`,`user_id`)
) ENGINE=InnoDB AUTO_INCREMENT=10000 DEFAULT CHARSET=utf8mb4 COMMENT='企业微信成员表';

-- [CHECK POINT] --

-- 企业微信成员与部门关联表：成员与部门多对多关系，含按部门负责人标记
CREATE TABLE `gaz_wecom_member_department` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `store_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '企业ID(租户)',
  `user_id` varchar(64) NOT NULL DEFAULT '' COMMENT '企微成员UserID(关联gaz_wecom_member.user_id)',
  `department_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '部门ID(关联gaz_wecom_department.department_id)',
  `is_leader` tinyint(3) unsigned NOT NULL DEFAULT '0' COMMENT '是否该部门负责人(1是 0否)',
  `sort` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '成员在部门内的排序',
  `created_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_member` (`store_id`,`user_id`),
  KEY `idx_department` (`store_id`,`department_id`)
) ENGINE=InnoDB AUTO_INCREMENT=10000 DEFAULT CHARSET=utf8mb4 COMMENT='企业微信成员与部门关联表';

-- [CHECK POINT] --

-- 企业微信外部联系人(SCRM客户)表：对齐企微客户详情接口external_contact
CREATE TABLE `gaz_wecom_customer` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `store_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '企业ID(租户)',
  `external_userid` varchar(64) NOT NULL DEFAULT '' COMMENT '企微外部联系人UserID',
  `name` varchar(64) NOT NULL DEFAULT '' COMMENT '客户名称(昵称)',
  `position` varchar(128) NOT NULL DEFAULT '' COMMENT '职位',
  `avatar` varchar(255) NOT NULL DEFAULT '' COMMENT '头像URL',
  `corp_name` varchar(128) NOT NULL DEFAULT '' COMMENT '公司简称',
  `corp_full_name` varchar(255) NOT NULL DEFAULT '' COMMENT '公司全称',
  `type` tinyint(3) unsigned NOT NULL DEFAULT '1' COMMENT '类型(1微信用户 2企业微信用户)',
  `gender` tinyint(3) unsigned NOT NULL DEFAULT '0' COMMENT '性别(0未知 1男 2女)',
  `unionid` varchar(64) NOT NULL DEFAULT '' COMMENT '微信UnionID',
  `created_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '创建时间',
  `updated_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '更新时间',
  `deleted_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_store_customer` (`store_id`,`external_userid`),
  KEY `idx_store_unionid` (`store_id`,`unionid`)
) ENGINE=InnoDB AUTO_INCREMENT=10000 DEFAULT CHARSET=utf8mb4 COMMENT='企业微信外部联系人(客户)表';

-- [CHECK POINT] --

-- 企业微信客户跟进人关系表：对齐企微客户详情接口follow_user，一个客户可被多个成员跟进
CREATE TABLE `gaz_wecom_customer_follow` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `store_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '企业ID(租户)',
  `external_userid` varchar(64) NOT NULL DEFAULT '' COMMENT '客户UserID(关联gaz_wecom_customer)',
  `user_id` varchar(64) NOT NULL DEFAULT '' COMMENT '跟进人企微UserID',
  `remark` varchar(128) NOT NULL DEFAULT '' COMMENT '跟进人备注',
  `description` varchar(255) NOT NULL DEFAULT '' COMMENT '跟进人描述',
  `remark_corp_name` varchar(128) NOT NULL DEFAULT '' COMMENT '备注公司名',
  `remark_mobiles` varchar(255) NOT NULL DEFAULT '' COMMENT '备注手机号(逗号分隔)',
  `add_way` tinyint(3) unsigned NOT NULL DEFAULT '0' COMMENT '添加方式(与企微add_way一致)',
  `state` varchar(30) NOT NULL DEFAULT '' COMMENT '添加渠道标识',
  `create_time` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '添加客户时间',
  `created_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '创建时间',
  `updated_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '更新时间',
  `deleted_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_store_follow` (`store_id`,`external_userid`,`user_id`),
  KEY `idx_store_user` (`store_id`,`user_id`),
  KEY `idx_store_state` (`store_id`,`state`)
) ENGINE=InnoDB AUTO_INCREMENT=10000 DEFAULT CHARSET=utf8mb4 COMMENT='企业微信客户跟进人关系表';

-- [CHECK POINT] --

-- 企业微信客户群表：对齐企微客户群列表/详情接口group_chat
CREATE TABLE `gaz_wecom_group` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `store_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '企业ID(租户)',
  `chat_id` varchar(64) NOT NULL DEFAULT '' COMMENT '客户群ID',
  `name` varchar(255) NOT NULL DEFAULT '' COMMENT '群名',
  `owner` varchar(64) NOT NULL DEFAULT '' COMMENT '群主企微UserID',
  `create_time` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '群创建时间',
  `notice` varchar(2000) NOT NULL DEFAULT '' COMMENT '群公告',
  `status` tinyint(3) unsigned NOT NULL DEFAULT '0' COMMENT '群状态(0跟进人正常 1跟进人离职 2离职继承中 3离职继承完成)',
  `member_count` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '群成员数(冗余, 同步时按member_list长度写入)',
  `member_version` varchar(64) NOT NULL DEFAULT '' COMMENT '成员版本号(详情接口返回, 用于增量同步)',
  `created_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '创建时间',
  `updated_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '更新时间',
  `deleted_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_store_group` (`store_id`,`chat_id`),
  KEY `idx_store_owner` (`store_id`,`owner`)
) ENGINE=InnoDB AUTO_INCREMENT=10000 DEFAULT CHARSET=utf8mb4 COMMENT='企业微信客户群表';

-- [CHECK POINT] --

-- 企业微信客户群成员表：对齐企微客户群详情接口member_list，群与成员多对多
CREATE TABLE `gaz_wecom_group_member` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `store_id` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '企业ID(租户)',
  `chat_id` varchar(64) NOT NULL DEFAULT '' COMMENT '客户群ID(关联gaz_wecom_group)',
  `user_id` varchar(64) NOT NULL DEFAULT '' COMMENT '成员企微UserID(外部联系人为external_userid)',
  `type` tinyint(3) unsigned NOT NULL DEFAULT '1' COMMENT '成员类型(1企业成员 2外部联系人)',
  `unionid` varchar(64) NOT NULL DEFAULT '' COMMENT '微信UnionID(外部联系人)',
  `join_time` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '入群时间',
  `join_scene` tinyint(3) unsigned NOT NULL DEFAULT '0' COMMENT '入群方式(1成员直接邀请 2邀请链接 3扫描群二维码)',
  `invitor_userid` varchar(64) NOT NULL DEFAULT '' COMMENT '邀请人企微UserID(外部联系人入群时)',
  `group_nickname` varchar(64) NOT NULL DEFAULT '' COMMENT '群内昵称',
  `name` varchar(64) NOT NULL DEFAULT '' COMMENT '成员名称(外部联系人为微信昵称)',
  `is_admin` tinyint(3) unsigned NOT NULL DEFAULT '0' COMMENT '是否群管理员(1是 0否, 对齐admin_list)',
  `created_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '创建时间',
  `deleted_at` int(11) unsigned NOT NULL DEFAULT '0' COMMENT '删除时间',
  PRIMARY KEY (`id`),
  KEY `idx_store_member` (`store_id`,`chat_id`),
  KEY `idx_store_user` (`store_id`,`user_id`)
) ENGINE=InnoDB AUTO_INCREMENT=10000 DEFAULT CHARSET=utf8mb4 COMMENT='企业微信客户群成员表';
