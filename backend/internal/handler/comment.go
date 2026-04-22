// 评论处理器
// 文件名: comment.go
// 路径: /workspace/backend/internal/handler/comment.go

package handler

import (
	"net/http"
	"strconv"

	"blog-cms/internal/model"
	"blog-cms/internal/service"

	"github.com/gin-gonic/gin"
)

type CommentHandler struct {
	commentService *service.CommentService
}

func NewCommentHandler() *CommentHandler {
	return &CommentHandler{
		commentService: service.GetCommentService(),
	}
}

// GetComments 获取文章评论列表
func (h *CommentHandler) GetComments(c *gin.Context) {
	articleIDStr := c.Param("article_id")
	articleID, err := strconv.ParseUint(articleIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的文章ID",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	comments, total, err := h.commentService.GetCommentsByArticleID(articleID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取评论失败：" + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"list":      comments,
			"total":     total,
			"page":      page,
			"pageSize":  pageSize,
		},
		"message": "success",
	})
}

// CreateComment 创建评论
func (h *CommentHandler) CreateComment(c *gin.Context) {
	articleIDStr := c.Param("article_id")
	articleID, err := strconv.ParseUint(articleIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的文章ID",
		})
		return
	}

	var comment model.Comment
	if err := c.ShouldBindJSON(&comment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误：" + err.Error(),
		})
		return
	}

	// 设置文章ID
	comment.ArticleID = articleID

	// 获取客户端IP
	clientIP := c.ClientIP()

	if err := h.commentService.CreateComment(&comment, clientIP); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "发表评论失败：" + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"data":    comment,
		"message": "评论成功",
	})
}

// DeleteComment 删除评论 (管理员)
func (h *CommentHandler) DeleteComment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的评论ID",
		})
		return
	}

	if err := h.commentService.DeleteComment(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除评论失败：" + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "评论删除成功",
	})
}

// AuditComment 审核评论 (管理员)
func (h *CommentHandler) AuditComment(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的评论ID",
		})
		return
	}

	var req struct {
		Status int `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
		})
		return
	}

	if err := h.commentService.AuditComment(id, req.Status); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "审核失败：" + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "审核成功",
	})
}
