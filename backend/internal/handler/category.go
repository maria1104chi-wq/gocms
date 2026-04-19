// 分类处理器
// 文件名: category.go
// 路径: /workspace/backend/internal/handler/category.go

package handler

import (
	"net/http"

	"blog-cms/internal/database"
	"blog-cms/internal/model"

	"github.com/gin-gonic/gin"
)

type CategoryHandler struct{}

func NewCategoryHandler() *CategoryHandler {
	return &CategoryHandler{}
}

// GetCategories 获取所有分类
func (h *CategoryHandler) GetCategories(c *gin.Context) {
	db := database.GetDB()
	var categories []model.Category
	
	if err := db.Order("sort ASC, id ASC").Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "获取分类失败：" + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"data":    categories,
		"message": "success",
	})
}
