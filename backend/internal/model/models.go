// 数据库模型定义
// 文件名: models.go
// 路径: /workspace/backend/internal/model/models.go

package model

import (
	"time"
)

// User 用户模型
type User struct {
	ID           uint64    `gorm:"primaryKey" json:"id"`
	Username     string    `gorm:"size:50;uniqueIndex;not null" json:"username"`
	PasswordHash string    `gorm:"size:255;not null" json:"-"`
	Phone        string    `gorm:"size:20;uniqueIndex" json:"phone"`
	Role         int       `gorm:"default:1" json:"role"` // 1=普通用户, 2=版块管理员, 3=系统管理员
	Status       int       `gorm:"default:1" json:"status"`
	Avatar       string    `gorm:"size:255" json:"avatar"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Category 分类模型
type Category struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:50;not null" json:"name"`
	Slug        string    `gorm:"size:50;uniqueIndex;not null" json:"slug"`
	Description string    `gorm:"size:255" json:"description"`
	Sort        int       `gorm:"default:0" json:"sort"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// UserCategory 用户-分类关联模型
type UserCategory struct {
	ID         uint64    `gorm:"primaryKey" json:"id"`
	UserID     uint64    `gorm:"uniqueIndex:uk_user_category;not null" json:"user_id"`
	CategoryID uint64    `gorm:"uniqueIndex:uk_user_category;not null" json:"category_id"`
	CreatedAt  time.Time `json:"created_at"`
}

// Article 文章模型
type Article struct {
	ID              uint64    `gorm:"primaryKey" json:"id"`
	Title           string    `gorm:"size:200;not null" json:"title"`
	Slug            string    `gorm:"size:100;uniqueIndex" json:"slug"`
	Summary         string    `gorm:"size:500" json:"summary"`
	Content         string    `gorm:"type:text;not null" json:"content"`
	CategoryID      uint64    `gorm:"not null;index" json:"category_id"`
	AuthorID        uint64    `gorm:"not null;index" json:"author_id"`
	CoverImage      string    `gorm:"size:255" json:"cover_image"`
	ViewCount       uint64    `gorm:"default:0" json:"view_count"`
	LikeCount       uint64    `gorm:"default:0" json:"like_count"`
	ShareCount      uint64    `gorm:"default:0" json:"share_count"`
	CommentCount    uint64    `gorm:"default:0" json:"comment_count"`
	Status          int       `gorm:"default:1;index" json:"status"` // 0=草稿, 1=发布, 2=下架
	IsTop           int       `gorm:"default:0" json:"is_top"`
	SeoKeywords     string    `gorm:"size:255" json:"seo_keywords"`
	SeoDescription  string    `gorm:"size:500" json:"seo_description"`
	CreatedAt       time.Time `gorm:"index" json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	PublishedAt     *time.Time `json:"published_at"`
	
	// 关联字段
	Category *Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Author   *User     `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
}

// Comment 评论模型
type Comment struct {
	ID         uint64    `gorm:"primaryKey" json:"id"`
	ArticleID  uint64    `gorm:"not null;index" json:"article_id"`
	UserID     *uint64   `json:"user_id"` // NULL表示匿名
	ParentID   *uint64   `gorm:"index" json:"parent_id"` // NULL表示主评论
	Content    string    `gorm:"type:text;not null" json:"content"`
	IPAddress  string    `gorm:"size:45;not null" json:"ip_address"`
	IPLocation string    `gorm:"size:100" json:"ip_location"`
	Status     int       `gorm:"default:1" json:"status"` // 0=待审核, 1=显示, 2=隐藏
	CreatedAt  time.Time `gorm:"index" json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	
	// 关联字段
	User     *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Parent   *Comment  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Replies  []Comment `gorm:"foreignKey:ParentID" json:"replies,omitempty"`
}

// SensitiveWord 敏感词模型
type SensitiveWord struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	Word      string    `gorm:"size:100;uniqueIndex;not null" json:"word"`
	Category  string    `gorm:"size:50;default:general" json:"category"`
	CreatedAt time.Time `json:"created_at"`
}

// SMSCode 短信验证码模型
type SMSCode struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	Phone     string    `gorm:"size:20;not null;index:idx_phone_expires" json:"phone"`
	Code      string    `gorm:"size:10;not null" json:"code"`
	ExpiresAt time.Time `gorm:"not null;index:idx_phone_expires" json:"expires_at"`
	Used      int       `gorm:"default:0" json:"used"`
	CreatedAt time.Time `json:"created_at"`
}

// VisitLog 访问日志模型
type VisitLog struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	IPAddress string    `gorm:"size:45;not null;index" json:"ip_address"`
	IPLocation string   `gorm:"size:100" json:"ip_location"`
	URL       string    `gorm:"size:255;not null" json:"url"`
	Method    string    `gorm:"size:10;not null" json:"method"`
	UserAgent string    `gorm:"size:500" json:"user_agent"`
	Refer     string    `gorm:"size:255" json:"refer"`
	ArticleID *uint64   `gorm:"index" json:"article_id"`
	CreatedAt time.Time `gorm:"index" json:"created_at"`
}

// Like 点赞记录模型
type Like struct {
	ID        uint64    `gorm:"primaryKey" json:"id"`
	ArticleID uint64    `gorm:"not null;uniqueIndex:uk_article_user;uniqueIndex:uk_article_ip" json:"article_id"`
	UserID    *uint64   `gorm:"uniqueIndex:uk_article_user" json:"user_id"` // NULL表示匿名
	IPAddress string    `gorm:"size:45;not null;uniqueIndex:uk_article_ip" json:"ip_address"`
	CreatedAt time.Time `json:"created_at"`
}

// SystemConfig 系统配置模型
type SystemConfig struct {
	ID          uint64    `gorm:"primaryKey" json:"id"`
	ConfigKey   string    `gorm:"size:50;uniqueIndex;not null" json:"config_key"`
	ConfigValue string    `gorm:"type:text" json:"config_value"`
	Description string    `gorm:"size:255" json:"description"`
	UpdatedAt   time.Time `json:"updated_at"`
}
