// 主程序入口
// 文件名: main.go
// 路径: /workspace/backend/cmd/main.go

package main

import (
	"log"
	"os"
	"time"

	"blog-cms/internal/config"
	"blog-cms/internal/database"
	"blog-cms/internal/handler"
	"blog-cms/internal/middleware"
	"blog-cms/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 初始化数据库
	if err := database.InitDB(&cfg.Database); err != nil {
		log.Fatalf("数据库初始化失败：%v", err)
	}

	// 初始化Redis
	if err := database.InitRedis(&cfg.Redis); err != nil {
		log.Fatalf("Redis初始化失败：%v", err)
	}

	// 初始化JWT
	middleware.InitJWT(&cfg.JWT)

	// 加载敏感词到内存
	sensitiveService := service.GetSensitiveService()
	if err := sensitiveService.LoadWords(); err != nil {
		log.Printf("警告：敏感词加载失败：%v", err)
	} else {
		log.Println("敏感词库加载成功")
	}

	// 设置Gin模式
	gin.SetMode(cfg.Server.Mode)

	// 创建Gin引擎
	r := gin.Default()

	// 配置CORS (跨域)
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// 静态文件服务 (上传的文件)
	r.Static("/uploads", "./static/uploads")

	// API路由组
	api := r.Group("/api")
	{
		// 文章相关路由
		articleHandler := handler.NewArticleHandler()
		articles := api.Group("/articles")
		{
			articles.GET("", articleHandler.GetArticleList)           // 获取文章列表
			articles.GET("/top", articleHandler.GetTopArticles)       // 获取热门文章
			articles.GET("/:slug", articleHandler.GetArticleDetail)   // 获取文章详情 (伪静态)
			articles.POST("/:id/like", articleHandler.LikeArticle)    // 点赞文章
			articles.POST("/:id/share", articleHandler.ShareArticle)  // 分享文章
			
			// 需要认证的路由
			authArticles := articles.Group("")
			authArticles.Use(middleware.JWTAuth())
			{
				authArticles.POST("", articleHandler.CreateArticle)     // 创建文章
				authArticles.PUT("/:id", articleHandler.UpdateArticle)  // 更新文章
				authArticles.DELETE("/:id", articleHandler.DeleteArticle) // 删除文章
			}
		}

		// 评论相关路由
		commentHandler := handler.NewCommentHandler()
		comments := api.Group("/comments")
		{
			comments.GET("/article/:article_id", commentHandler.GetComments)      // 获取文章评论
			comments.POST("/article/:article_id", commentHandler.CreateComment)   // 创建评论
			
			// 需要管理员权限的路由
			adminComments := comments.Group("")
			adminComments.Use(middleware.JWTAuth(), middleware.RequireRole(2))
			{
				adminComments.DELETE("/:id", commentHandler.DeleteComment)    // 删除评论
				adminComments.POST("/:id/audit", commentHandler.AuditComment) // 审核评论
			}
		}

		// 分类相关路由
		categoryHandler := handler.NewCategoryHandler()
		categories := api.Group("/categories")
		{
			categories.GET("", categoryHandler.GetCategories) // 获取所有分类
		}

		// 用户相关路由
		userHandler := handler.NewUserHandler()
		users := api.Group("/users")
		{
			users.POST("/register", userHandler.Register)         // 注册
			users.POST("/login", userHandler.Login)               // 登录
			users.POST("/sms/send", userHandler.SendSMSCode)      // 发送短信验证码
			users.POST("/sms/verify", userHandler.VerifySMSCode)  // 验证短信验证码
			users.GET("/profile", middleware.JWTAuth(), userHandler.GetProfile) // 获取个人信息
		}

		// 管理后台路由 (需要系统管理员权限)
		admin := api.Group("/admin")
		admin.Use(middleware.JWTAuth(), middleware.RequireRole(3))
		{
			// 用户管理
			admin.GET("/users", userHandler.GetUsers)        // 获取用户列表
			admin.PUT("/users/:id/role", userHandler.UpdateUserRole) // 更新用户角色
			
			// 敏感词管理
			sensitiveGroup := admin.Group("/sensitive")
			{
				sensitiveGroup.GET("", getSensitiveWords)     // 获取敏感词列表
				sensitiveGroup.POST("", addSensitiveWord)     // 添加敏感词
				sensitiveGroup.DELETE("/:word", deleteSensitiveWord) // 删除敏感词
			}
			
			// 统计信息
			admin.GET("/stats", getStats) // 获取统计数据
		}
	}

	// 启动服务器
	port := ":" + cfg.Server.Port
	log.Printf("服务器启动在端口 %s", port)
	
	if err := r.Run(port); err != nil {
		log.Fatalf("服务器启动失败：%v", err)
	}
}

// getSensitiveWords 获取敏感词列表 (管理后台)
func getSensitiveWords(c *gin.Context) {
	// TODO: 实现获取敏感词列表逻辑
	c.JSON(200, gin.H{"code": 0, "data": []string{}, "message": "success"})
}

// addSensitiveWord 添加敏感词 (管理后台)
func addSensitiveWord(c *gin.Context) {
	var req struct {
		Word     string `json:"word"`
		Category string `json:"category"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"code": 400, "message": "参数错误"})
		return
	}
	
	sensitiveService := service.GetSensitiveService()
	if err := sensitiveService.AddWord(req.Word, req.Category); err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "添加失败：" + err.Error()})
		return
	}
	
	c.JSON(200, gin.H{"code": 0, "message": "添加成功"})
}

// deleteSensitiveWord 删除敏感词 (管理后台)
func deleteSensitiveWord(c *gin.Context) {
	word := c.Param("word")
	
	sensitiveService := service.GetSensitiveService()
	if err := sensitiveService.DeleteWord(word); err != nil {
		c.JSON(500, gin.H{"code": 500, "message": "删除失败：" + err.Error()})
		return
	}
	
	c.JSON(200, gin.H{"code": 0, "message": "删除成功"})
}

// getStats 获取统计数据 (管理后台)
func getStats(c *gin.Context) {
	// TODO: 实现统计数据查询逻辑
	c.JSON(200, gin.H{
		"code": 0,
		"data": gin.H{
			"user_count":     0,
			"article_count":  0,
			"comment_count":  0,
			"view_count":     0,
			"today_views":    0,
		},
		"message": "success",
	})
}
