-- 数据库初始化脚本
-- 文件名: deploy/init.sql
-- 路径: /workspace/deploy/init.sql

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- 创建数据库
CREATE DATABASE IF NOT EXISTS `blog_cms` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE `blog_cms`;

-- 1. 用户表 (users)
DROP TABLE IF EXISTS `users`;
CREATE TABLE `users` (
  `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '用户ID',
  `username` varchar(50) NOT NULL COMMENT '用户名',
  `password_hash` varchar(255) NOT NULL COMMENT '密码哈希',
  `phone` varchar(20) DEFAULT NULL COMMENT '手机号码',
  `role` tinyint(4) NOT NULL DEFAULT 1 COMMENT '角色: 1=普通用户, 2=版块管理员, 3=系统管理员',
  `status` tinyint(4) NOT NULL DEFAULT 1 COMMENT '状态: 0=禁用, 1=正常',
  `avatar` varchar(255) DEFAULT NULL COMMENT '头像URL',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_username` (`username`),
  UNIQUE KEY `uk_phone` (`phone`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

-- 2. 版块分类表 (categories)
DROP TABLE IF EXISTS `categories`;
CREATE TABLE `categories` (
  `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '分类ID',
  `name` varchar(50) NOT NULL COMMENT '分类名称',
  `slug` varchar(50) NOT NULL COMMENT 'URL标识',
  `description` varchar(255) DEFAULT NULL COMMENT '描述',
  `sort` int(11) NOT NULL DEFAULT 0 COMMENT '排序',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_slug` (`slug`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='版块分类表';

-- 3. 用户-版块关联表 (user_categories) - 多对多关系
DROP TABLE IF EXISTS `user_categories`;
CREATE TABLE `user_categories` (
  `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `user_id` bigint(20) UNSIGNED NOT NULL COMMENT '用户ID',
  `category_id` bigint(20) UNSIGNED NOT NULL COMMENT '分类ID',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '授权时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_user_category` (`user_id`, `category_id`),
  CONSTRAINT `fk_uc_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_uc_category` FOREIGN KEY (`category_id`) REFERENCES `categories` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户版块权限关联表';

-- 4. 文章表 (articles)
DROP TABLE IF EXISTS `articles`;
CREATE TABLE `articles` (
  `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '文章ID',
  `title` varchar(200) NOT NULL COMMENT '标题',
  `slug` varchar(100) DEFAULT NULL COMMENT 'URL标识 (伪静态)',
  `summary` varchar(500) DEFAULT NULL COMMENT '摘要',
  `content` text NOT NULL COMMENT '内容 (Markdown/HTML)',
  `category_id` bigint(20) UNSIGNED NOT NULL COMMENT '分类ID',
  `author_id` bigint(20) UNSIGNED NOT NULL COMMENT '作者ID',
  `cover_image` varchar(255) DEFAULT NULL COMMENT '封面图URL',
  `view_count` bigint(20) UNSIGNED NOT NULL DEFAULT 0 COMMENT '点击数',
  `like_count` bigint(20) UNSIGNED NOT NULL DEFAULT 0 COMMENT '点赞数',
  `share_count` bigint(20) UNSIGNED NOT NULL DEFAULT 0 COMMENT '分享数',
  `comment_count` bigint(20) UNSIGNED NOT NULL DEFAULT 0 COMMENT '评论数',
  `status` tinyint(4) NOT NULL DEFAULT 1 COMMENT '状态: 0=草稿, 1=发布, 2=下架',
  `is_top` tinyint(4) NOT NULL DEFAULT 0 COMMENT '是否置顶: 0=否, 1=是',
  `seo_keywords` varchar(255) DEFAULT NULL COMMENT 'SEO关键词',
  `seo_description` varchar(500) DEFAULT NULL COMMENT 'SEO描述',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  `published_at` datetime DEFAULT NULL COMMENT '发布时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_slug` (`slug`),
  KEY `idx_category` (`category_id`),
  KEY `idx_author` (`author_id`),
  KEY `idx_status_created` (`status`, `created_at`),
  CONSTRAINT `fk_article_category` FOREIGN KEY (`category_id`) REFERENCES `categories` (`id`) ON DELETE RESTRICT,
  CONSTRAINT `fk_article_author` FOREIGN KEY (`author_id`) REFERENCES `users` (`id`) ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文章表';

-- 5. 评论表 (comments)
DROP TABLE IF EXISTS `comments`;
CREATE TABLE `comments` (
  `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '评论ID',
  `article_id` bigint(20) UNSIGNED NOT NULL COMMENT '文章ID',
  `user_id` bigint(20) UNSIGNED DEFAULT NULL COMMENT '用户ID (NULL表示匿名)',
  `parent_id` bigint(20) UNSIGNED DEFAULT NULL COMMENT '父评论ID (用于跟评)',
  `content` text NOT NULL COMMENT '评论内容',
  `ip_address` varchar(45) NOT NULL COMMENT 'IP地址',
  `ip_location` varchar(100) DEFAULT NULL COMMENT 'IP归属地',
  `status` tinyint(4) NOT NULL DEFAULT 1 COMMENT '状态: 0=待审核, 1=显示, 2=隐藏',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_article` (`article_id`),
  KEY `idx_parent` (`parent_id`),
  KEY `idx_created` (`created_at`),
  CONSTRAINT `fk_comment_article` FOREIGN KEY (`article_id`) REFERENCES `articles` (`id`) ON DELETE CASCADE,
  CONSTRAINT `fk_comment_user` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE SET NULL,
  CONSTRAINT `fk_comment_parent` FOREIGN KEY (`parent_id`) REFERENCES `comments` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='评论表';

-- 6. 敏感词库表 (sensitive_words)
DROP TABLE IF EXISTS `sensitive_words`;
CREATE TABLE `sensitive_words` (
  `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `word` varchar(100) NOT NULL COMMENT '敏感词',
  `category` varchar(50) DEFAULT 'general' COMMENT '分类',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_word` (`word`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='敏感词库表';

-- 7. 短信验证码表 (sms_codes)
DROP TABLE IF EXISTS `sms_codes`;
CREATE TABLE `sms_codes` (
  `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `phone` varchar(20) NOT NULL COMMENT '手机号',
  `code` varchar(10) NOT NULL COMMENT '验证码',
  `expires_at` datetime NOT NULL COMMENT '过期时间',
  `used` tinyint(4) NOT NULL DEFAULT 0 COMMENT '是否已使用: 0=否, 1=是',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_phone_expires` (`phone`, `expires_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='短信验证码表';

-- 8. 访问统计表 (visit_logs)
DROP TABLE IF EXISTS `visit_logs`;
CREATE TABLE `visit_logs` (
  `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `ip_address` varchar(45) NOT NULL COMMENT 'IP地址',
  `ip_location` varchar(100) DEFAULT NULL COMMENT 'IP归属地',
  `url` varchar(255) NOT NULL COMMENT '访问URL',
  `method` varchar(10) NOT NULL COMMENT '请求方法',
  `user_agent` varchar(500) DEFAULT NULL COMMENT 'User-Agent',
  `refer` varchar(255) DEFAULT NULL COMMENT '来源页面',
  `article_id` bigint(20) UNSIGNED DEFAULT NULL COMMENT '关联文章ID',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_ip` (`ip_address`),
  KEY `idx_article` (`article_id`),
  KEY `idx_created` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='访问日志统计表';

-- 9. 点赞记录表 (likes) - 防止重复点赞
DROP TABLE IF EXISTS `likes`;
CREATE TABLE `likes` (
  `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `article_id` bigint(20) UNSIGNED NOT NULL COMMENT '文章ID',
  `user_id` bigint(20) UNSIGNED DEFAULT NULL COMMENT '用户ID (NULL表示匿名)',
  `ip_address` varchar(45) NOT NULL COMMENT 'IP地址 (匿名时用)',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_article_user` (`article_id`, `user_id`),
  UNIQUE KEY `uk_article_ip` (`article_id`, `ip_address`, `user_id`),
  CONSTRAINT `fk_like_article` FOREIGN KEY (`article_id`) REFERENCES `articles` (`id`) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='点赞记录表';

-- 10. 系统配置表 (system_configs)
DROP TABLE IF EXISTS `system_configs`;
CREATE TABLE `system_configs` (
  `id` bigint(20) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT 'ID',
  `config_key` varchar(50) NOT NULL COMMENT '配置键',
  `config_value` text COMMENT '配置值',
  `description` varchar(255) DEFAULT NULL COMMENT '描述',
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_config_key` (`config_key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统配置表';

-- 插入初始数据
-- 默认管理员账号 (密码: admin123, 实际部署请修改)
INSERT INTO `users` (`username`, `password_hash`, `phone`, `role`, `status`) VALUES 
('admin', '$2a$10$X7VJkZzKqZzKqZzKqZzKqOexamplehashhere', '13800138000', 3, 1);

-- 默认分类
INSERT INTO `categories` (`name`, `slug`, `description`, `sort`) VALUES 
('技术前沿', 'tech', '最新技术动态', 1),
('生活随笔', 'life', '生活感悟', 2),
('开源项目', 'opensource', '开源项目分享', 3);

-- 插入部分敏感词示例
INSERT INTO `sensitive_words` (`word`, `category`) VALUES 
('敏感词1', 'political'),
('敏感词2', 'advertising'),
('广告', 'advertising');

SET FOREIGN_KEY_CHECKS = 1;
