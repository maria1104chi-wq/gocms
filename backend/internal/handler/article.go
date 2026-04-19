// 文章处理器
// 文件名: article.go
// 路径: /workspace/backend/internal/handler/article.go

package handler

import (
	"net/http"
	"strconv"

	"blog-cms/internal/middleware"
	"blog-cms/internal/model"
	"blog-cms/internal/service"

	"github.com/gin-gonic/gin"
)

type ArticleHandler struct {
	articleService *service.ArticleService
}

func NewArticleHandler() *ArticleHandler {
	return &ArticleHandler{
		articleService: service.GetArticleService(),
	}
}

// GetArticleList 获取文章列表
// @Summary 获取文章列表
// @Tags 文章
// @Accept json
// @Produce json
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Param category_id query int false "分类ID"
// @Param keyword query string false "搜索关键词"
// @Success 200 {object} Response
// @Router /api/articles [get]
func (h *ArticleHandler) GetArticleList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	var categoryID *uint64
	if cid := c.Query("category_id"); cid != "" {
		id, err := strconv.ParseUint(cid, 10, 64)
		if err == nil && id > 0 {
			categoryID = &id
		}
	}

	keyword := c.Query("keyword")

	articles, total, err := h.articleService.GetArticleList(page, pageSize, categoryID, keyword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取文章列表失败：" + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"list":      articles,
			"total":     total,
			"page":      page,
			"pageSize":  pageSize,
			"totalPage": (total + int64(pageSize) - 1) / int64(pageSize),
		},
		"message": "success",
	})
}

// GetArticleDetail 获取文章详情
// @Summary 获取文章详情
// @Tags 文章
// @Accept json
// @Produce json
// @Param slug path string true "文章标识"
// @Success 200 {object} Response
// @Router /api/articles/:slug [get]
func (h *ArticleHandler) GetArticleDetail(c *gin.Context) {
	slug := c.Param("slug")
	
	if slug == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "文章标识不能为空",
		})
		return
	}

	article, err := h.articleService.GetArticleBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "文章不存在",
		})
		return
	}

	// 异步增加点击数 (不阻塞响应)
	go func() {
		ip := c.ClientIP()
		h.articleService.IncrementViewCount(article.ID, ip)
	}()

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"data":    article,
		"message": "success",
	})
}

// CreateArticle 创建文章
// @Summary 创建文章
// @Tags 文章
// @Accept json
// @Produce json
// @Param article body model.Article true "文章信息"
// @Success 200 {object} Response
// @Security BearerAuth
// @Router /api/articles [post]
func (h *ArticleHandler) CreateArticle(c *gin.Context) {
	userID, username, role, ok := middleware.GetUserFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "请先登录",
		})
		return
	}

	// 检查权限 (至少需要版块管理员或系统管理员)
	if role < 2 {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "权限不足，无法发布文章",
		})
		return
	}

	var article model.Article
	if err := c.ShouldBindJSON(&article); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误：" + err.Error(),
		})
		return
	}

	// 设置作者
	article.AuthorID = userID

	if err := h.articleService.CreateArticle(&article); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "创建文章失败：" + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"data":    article,
		"message": "文章创建成功",
	})
}

// UpdateArticle 更新文章
// @Summary 更新文章
// @Tags 文章
// @Accept json
// @Produce json
// @Param id path int true "文章ID"
// @Param article body model.Article true "文章信息"
// @Success 200 {object} Response
// @Security BearerAuth
// @Router /api/articles/:id [put]
func (h *ArticleHandler) UpdateArticle(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的文章ID",
		})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误：" + err.Error(),
		})
		return
	}

	if err := h.articleService.UpdateArticle(id, updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新文章失败：" + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "文章更新成功",
	})
}

// DeleteArticle 删除文章
// @Summary 删除文章
// @Tags 文章
// @Accept json
// @Produce json
// @Param id path int true "文章ID"
// @Success 200 {object} Response
// @Security BearerAuth
// @Router /api/articles/:id [delete]
func (h *ArticleHandler) DeleteArticle(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的文章ID",
		})
		return
	}

	if err := h.articleService.DeleteArticle(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "删除文章失败：" + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "文章删除成功",
	})
}

// LikeArticle 点赞文章
// @Summary 点赞文章
// @Tags 文章
// @Accept json
// @Produce json
// @Param id path int true "文章ID"
// @Success 200 {object} Response
// @Router /api/articles/:id/like [post]
func (h *ArticleHandler) LikeArticle(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的文章ID",
		})
		return
	}

	ip := c.ClientIP()
	var userID *uint64
	
	// 尝试获取登录用户
	if uid, exists := c.Get("user_id"); exists {
		uidVal := uid.(uint64)
		userID = &uidVal
	}

	if err := h.articleService.LikeArticle(id, userID, ip); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "点赞成功",
	})
}

// ShareArticle 分享文章
// @Summary 分享文章
// @Tags 文章
// @Accept json
// @Produce json
// @Param id path int true "文章ID"
// @Success 200 {object} Response
// @Router /api/articles/:id/share [post]
func (h *ArticleHandler) ShareArticle(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的文章ID",
		})
		return
	}

	if err := h.articleService.ShareArticle(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "分享失败：" + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "分享成功",
	})
}

// GetTopArticles 获取热门文章
// @Summary 获取热门文章
// @Tags 文章
// @Accept json
// @Produce json
// @Param limit query int false "数量" default(10)
// @Success 200 {object} Response
// @Router /api/articles/top [get]
func (h *ArticleHandler) GetTopArticles(c *gin.Context) {
	limit := 10
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
		}
	}

	articles, err := h.articleService.GetTopArticles(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取热门文章失败：" + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"data":    articles,
		"message": "success",
	})
}
