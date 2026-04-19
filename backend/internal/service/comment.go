// 评论服务 - 包含敏感词过滤、IP归属地查询
// 文件名: comment.go
// 路径: /workspace/backend/internal/service/comment.go

package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"blog-cms/internal/database"
	"blog-cms/internal/model"
	"blog-cms/internal/utils"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

// CommentService 评论服务
type CommentService struct {
	db        *gorm.DB
	redis     *redis.Client
	sensitive *SensitiveService
}

var commentService *CommentService

// GetCommentService 获取评论服务单例
func GetCommentService() *CommentService {
	if commentService == nil {
		commentService = &CommentService{
			db:        database.GetDB(),
			redis:     database.GetRedis(),
			sensitive: GetSensitiveService(),
		}
	}
	return commentService
}

// CreateComment 创建评论 (事务性操作)
func (s *CommentService) CreateComment(comment *model.Comment, clientIP string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 检查文章是否存在
		var article model.Article
		if err := tx.First(&article, comment.ArticleID).Error; err != nil {
			return errors.New("文章不存在")
		}

		// 如果是跟评，检查父评论是否存在
		if comment.ParentID != nil {
			var parent model.Comment
			if err := tx.First(&parent, *comment.ParentID).Error; err != nil {
				return errors.New("父评论不存在")
			}
			// 确保父评论属于同一篇文章
			if parent.ArticleID != comment.ArticleID {
				return errors.New("父评论不属于该文章")
			}
		}

		// 设置IP地址
		comment.IPAddress = clientIP

		// 查询IP归属地
		comment.IPLocation = utils.GetIPLocation(clientIP)

		// 过滤敏感词
		comment.Content = s.sensitive.Filter(comment.Content)

		// 检查是否包含敏感词 (如果需要审核)
		// 这里简单处理：所有评论先显示，后续可扩展为敏感词多的需要审核
		comment.Status = 1 // 默认显示

		// 插入评论
		if err := tx.Create(comment).Error; err != nil {
			return err
		}

		// 原子增加文章评论数
		return tx.Model(&model.Article{}).
			Where("id = ?", comment.ArticleID).
			UpdateColumn("comment_count", gorm.Expr("comment_count + ?", 1)).Error
	})
}

// DeleteComment 删除评论 (事务性操作)
func (s *CommentService) DeleteComment(id uint64) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// 查找评论
		var comment model.Comment
		if err := tx.First(&comment, id).Error; err != nil {
			return errors.New("评论不存在")
		}

		// 删除评论 (子评论会因外键CASCADE自动删除)
		if err := tx.Delete(&comment).Error; err != nil {
			return err
		}

		// 原子减少文章评论数
		return tx.Model(&model.Article{}).
			Where("id = ?", comment.ArticleID).
			UpdateColumn("comment_count", gorm.Expr("CASE WHEN comment_count > 0 THEN comment_count - 1 ELSE 0 END")).Error
	})
}

// GetCommentsByArticleID 获取文章的评论列表 (支持分页)
func (s *CommentService) GetCommentsByArticleID(articleID uint64, page, pageSize int) ([]model.Comment, int64, error) {
	var comments []model.Comment
	var total int64

	query := s.db.Model(&model.Comment{}).
		Preload("User").
		Preload("Parent.User").
		Where("article_id = ? AND status = ?", articleID, 1)

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页排序 (按时间正序，方便查看对话)
	offset := (page - 1) * pageSize
	err := query.Order("created_at ASC").
		Limit(pageSize).Offset(offset).Find(&comments).Error

	return comments, total, err
}

// GetCommentTree 获取文章的评论树形结构
func (s *CommentService) GetCommentTree(articleID uint64) ([]model.Comment, error) {
	var allComments []model.Comment
	
	// 获取该文章的所有已审核评论
	err := s.db.Where("article_id = ? AND status = ?", articleID, 1).
		Order("created_at ASC").
		Find(&allComments).Error
	
	if err != nil {
		return nil, err
	}

	// 构建评论树
	commentMap := make(map[uint64]*model.Comment)
	var rootComments []model.Comment

	// 第一次遍历：建立映射并初始化Replies
	for i := range allComments {
		commentMap[allComments[i].ID] = &allComments[i]
		allComments[i].Replies = []model.Comment{}
	}

	// 第二次遍历：构建树形结构
	for _, comment := range allComments {
		if comment.ParentID == nil {
			// 根评论
			rootComments = append(rootComments, comment)
		} else {
			// 子评论，添加到父评论的Replies中
			if parent, ok := commentMap[*comment.ParentID]; ok {
				parent.Replies = append(parent.Replies, comment)
			}
		}
	}

	return rootComments, nil
}

// AuditComment 审核评论 (管理员使用)
func (s *CommentService) AuditComment(id uint64, status int) error {
	if status != 0 && status != 1 && status != 2 {
		return errors.New("无效的状态值")
	}

	result := s.db.Model(&model.Comment{}).
		Where("id = ?", id).
		Update("status", status)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return errors.New("评论不存在")
	}

	return nil
}

// GetRecentComments 获取最新评论 (用于侧栏或首页)
func (s *CommentService) GetRecentComments(limit int) ([]model.Comment, error) {
	var comments []model.Comment
	err := s.db.Where("status = ?", 1).
		Preload("User").
		Preload("Article").
		Order("created_at DESC").
		Limit(limit).
		Find(&comments).Error
	return comments, err
}

// ClearExpiredCache 清理过期的评论缓存 (定时任务调用)
func (s *CommentService) ClearExpiredCache(ctx context.Context) error {
	// 实现评论缓存清理逻辑
	// 可以使用Redis存储热门文章的评论缓存
	return nil
}

// GetIPLocation 获取IP归属地 (调用外部API)
// 注意：实际项目中建议使用本地IP库或购买商业IP库服务
func getIPLocation(ip string) string {
	// 简化实现：返回示例地理位置
	// 实际应调用 ip-api.com 或其他IP归属地服务
	ctx := context.Background()
	cacheKey := fmt.Sprintf("ip:location:%s", ip)

	// 尝试从Redis缓存获取
	redisClient := database.GetRedis()
	location, err := redisClient.Get(ctx, cacheKey).Result()
	if err == nil && location != "" {
		return location
	}

	// 调用外部API查询 (示例，实际需要实现HTTP请求)
	location = utils.GetIPLocation(ip)

	// 缓存结果 (24小时过期)
	if location != "" {
		redisClient.Set(ctx, cacheKey, location, 24*time.Hour)
	}

	return location
}
