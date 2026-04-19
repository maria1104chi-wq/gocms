// 文章服务 - 包含事务、原子计数、幂等性处理
// 文件名: article.go
// 路径: /workspace/backend/internal/service/article.go

package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"blog-cms/internal/database"
	"blog-cms/internal/model"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// ArticleService 文章服务
type ArticleService struct {
	db         *gorm.DB
	redis      *redis.Client
	sensitive  *SensitiveService
	viewMu     sync.Mutex // 防止并发点击数更新问题
}

var articleService *ArticleService

// GetArticleService 获取文章服务单例
func GetArticleService() *ArticleService {
	if articleService == nil {
		articleService = &ArticleService{
			db:        database.GetDB(),
			redis:     database.GetRedis(),
			sensitive: GetSensitiveService(),
		}
	}
	return articleService
}

// CreateArticle 创建文章 (事务性操作)
func (s *ArticleService) CreateArticle(article *model.Article) error {
	// 使用事务确保数据一致性
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 生成slug (如果未提供)
		if article.Slug == "" {
			article.Slug = generateSlug(article.Title)
		}

		// 检查slug是否已存在
		var existing model.Article
		if err := tx.Where("slug = ?", article.Slug).First(&existing).Error; err == nil {
			return errors.New("该URL标识已存在，请使用其他标题或手动指定slug")
		}

		// 过滤敏感词
		article.Title = s.sensitive.Filter(article.Title)
		article.Content = s.sensitive.Filter(article.Content)
		if article.Summary != "" {
			article.Summary = s.sensitive.Filter(article.Summary)
		}

		// 自动生成SEO信息
		if article.SeoKeywords == "" {
			article.SeoKeywords = article.Title
		}
		if article.SeoDescription == "" {
			// 截取内容前200字符作为描述
			desc := article.Content
			if len(desc) > 200 {
				desc = desc[:200]
			}
			article.SeoDescription = desc
		}

		// 设置发布时间
		now := time.Now()
		article.PublishedAt = &now

		// 插入文章
		if err := tx.Create(article).Error; err != nil {
			return err
		}

		// 更新分类的文章计数 (可选扩展)
		return nil
	})
}

// UpdateArticle 更新文章 (事务性操作)
func (s *ArticleService) UpdateArticle(id uint64, updates map[string]interface{}) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 过滤敏感词
		if title, ok := updates["title"].(string); ok {
			updates["title"] = s.sensitive.Filter(title)
		}
		if content, ok := updates["content"].(string); ok {
			updates["content"] = s.sensitive.Filter(content)
		}
		if summary, ok := updates["summary"].(string); ok && summary != "" {
			updates["summary"] = s.sensitive.Filter(summary)
		}

		// 更新文章
		result := tx.Model(&model.Article{}).Where("id = ?", id).Updates(updates)
		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return errors.New("文章不存在或未被修改")
		}

		return nil
	})
}

// DeleteArticle 删除文章 (事务性操作)
func (s *ArticleService) DeleteArticle(id uint64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 删除文章 (关联的评论会因外键CASCADE自动删除)
		result := tx.Delete(&model.Article{}, id)
		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return errors.New("文章不存在")
		}

		return nil
	})
}

// IncrementViewCount 增加点击数 (原子操作 + Redis缓存防刷)
func (s *ArticleService) IncrementViewCount(articleID uint64, ip string) error {
	s.viewMu.Lock()
	defer s.viewMu.Unlock()

	ctx := context.Background()
	cacheKey := fmt.Sprintf("article:view:%d:%s", articleID, ip)

	// 检查该IP是否已经点击过 (1小时内不能重复点击)
	exists, err := s.redis.Exists(ctx, cacheKey).Result()
	if err != nil {
		// Redis错误时降级处理，直接增加数据库计数
		return s.incrementViewCountDB(articleID)
	}

	if exists == 1 {
		// 该IP已经点击过，不重复计数
		return nil
	}

	// 标记该IP已点击 (1小时过期)
	err = s.redis.Set(ctx, cacheKey, "1", 1*time.Hour).Err()
	if err != nil {
		// Redis错误时降级处理
		return s.incrementViewCountDB(articleID)
	}

	// 增加数据库计数 (原子操作)
	return s.incrementViewCountDB(articleID)
}

// incrementViewCountDB 数据库层面原子增加点击数
func (s *ArticleService) incrementViewCountDB(articleID uint64) error {
	return s.db.Model(&model.Article{}).
		Where("id = ?", articleID).
		UpdateColumn("view_count", gorm.Expr("view_count + ?", 1)).Error
}

// LikeArticle 点赞文章 (幂等性处理)
func (s *ArticleService) LikeArticle(articleID uint64, userID *uint64, ip string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 检查是否已经点赞过
		var like model.Like
		query := tx.Where("article_id = ?", articleID)
		
		if userID != nil {
			// 登录用户
			query = query.Where("user_id = ?", *userID)
		} else {
			// 匿名用户
			query = query.Where("ip_address = ? AND user_id IS NULL", ip)
		}

		if err := query.First(&like).Error; err == nil {
			// 已经点赞过，返回错误 (幂等性)
			return errors.New("您已经点赞过该文章")
		}

		// 创建点赞记录
		newLike := model.Like{
			ArticleID: articleID,
			UserID:    userID,
			IPAddress: ip,
		}

		if err := tx.Create(&newLike).Error; err != nil {
			return err
		}

		// 原子增加文章点赞数
		return tx.Model(&model.Article{}).
			Where("id = ?", articleID).
			UpdateColumn("like_count", gorm.Expr("like_count + ?", 1)).Error
	})
}

// ShareArticle 分享文章
func (s *ArticleService) ShareArticle(articleID uint64) error {
	// 原子增加分享数
	return s.db.Model(&model.Article{}).
		Where("id = ?", articleID).
		UpdateColumn("share_count", gorm.Expr("share_count + ?", 1)).Error
}

// GetArticleByID 根据ID获取文章详情
func (s *ArticleService) GetArticleByID(id uint64) (*model.Article, error) {
	var article model.Article
	err := s.db.Preload("Category").Preload("Author").First(&article, id).Error
	if err != nil {
		return nil, err
	}
	return &article, nil
}

// GetArticleBySlug 根据slug获取文章详情 (伪静态URL支持)
func (s *ArticleService) GetArticleBySlug(slug string) (*model.Article, error) {
	var article model.Article
	err := s.db.Preload("Category").Preload("Author").Where("slug = ?", slug).First(&article).Error
	if err != nil {
		return nil, err
	}
	return &article, nil
}

// GetArticleList 获取文章列表 (支持分页、分类筛选、搜索)
func (s *ArticleService) GetArticleList(page, pageSize int, categoryID *uint64, keyword string) ([]model.Article, int64, error) {
	var articles []model.Article
	var total int64

	query := s.db.Model(&model.Article{}).Preload("Category").Preload("Author")

	// 只查询已发布的文章
	query = query.Where("status = ?", 1)

	// 分类筛选
	if categoryID != nil && *categoryID > 0 {
		query = query.Where("category_id = ?", *categoryID)
	}

	// 关键词搜索 (标题或内容)
	if keyword != "" {
		searchPattern := "%" + keyword + "%"
		query = query.Where("title LIKE ? OR content LIKE ?", searchPattern, searchPattern)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页排序 (置顶优先，然后按时间倒序)
	offset := (page - 1) * pageSize
	err := query.Order("is_top DESC").Order("created_at DESC").
		Limit(pageSize).Offset(offset).Find(&articles).Error

	return articles, total, err
}

// GetTopArticles 获取点击率排行前十的文章
func (s *ArticleService) GetTopArticles(limit int) ([]model.Article, error) {
	var articles []model.Article
	err := s.db.Where("status = ?", 1).
		Order("view_count DESC").
		Limit(limit).
		Find(&articles).Error
	return articles, err
}

// generateSlug 从标题生成URL友好的slug
func generateSlug(title string) string {
	// 简单实现：移除特殊字符，替换空格为连字符
	slug := strings.TrimSpace(title)
	slug = strings.ToLower(slug)
	
	// 移除中文标点等特殊字符 (简化处理)
	replacer := strings.NewReplacer(
		" ", "-",
		"?", "",
		"!", "",
		",", "",
		".", "",
	)
	slug = replacer.Replace(slug)

	// 添加时间戳保证唯一性
	slug = fmt.Sprintf("%s-%d", slug, time.Now().UnixNano())

	return slug
}
