// 用户处理器 - 包含注册、登录、短信验证等
// 文件名: user.go
// 路径: /workspace/backend/internal/handler/user.go

package handler

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"blog-cms/internal/database"
	"blog-cms/internal/middleware"
	"blog-cms/internal/model"
	"blog-cms/internal/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserHandler struct{}

func NewUserHandler() *UserHandler {
	return &UserHandler{}
}

// Register 用户注册
func (h *UserHandler) Register(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required,min=6"`
		Phone    string `json:"phone" binding:"required"`
		SMSCode  string `json:"sms_code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误：" + err.Error(),
		})
		return
	}

	db := database.GetDB()

	// 验证手机号格式
	if !utils.IsValidPhone(req.Phone) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "手机号格式不正确",
		})
		return
	}

	// 验证短信验证码
	if !verifySMSCode(db, req.Phone, req.SMSCode) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "短信验证码错误或已过期",
		})
		return
	}

	// 检查用户名是否已存在
	var existingUser model.User
	if err := db.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "用户名已存在",
		})
		return
	}

	// 检查手机号是否已存在
	if err := db.Where("phone = ?", req.Phone).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "手机号已被注册",
		})
		return
	}

	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "密码加密失败",
		})
		return
	}

	// 创建用户
	user := model.User{
		Username:     req.Username,
		PasswordHash: string(hashedPassword),
		Phone:        req.Phone,
		Role:         1, // 普通用户
		Status:       1,
	}

	if err := db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "注册失败：" + err.Error(),
		})
		return
	}

	// 生成JWT令牌
	token, err := middleware.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "生成令牌失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"user":  user,
			"token": token,
		},
		"message": "注册成功",
	})
}

// Login 用户登录
func (h *UserHandler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
		Phone    string `json:"phone"`           // 可选，用于短信验证登录
		SMSCode  string `json:"sms_code"`        // 可选，短信验证码
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误：" + err.Error(),
		})
		return
	}

	db := database.GetDB()
	var user model.User

	// 判断是密码登录还是短信登录
	if req.SMSCode != "" && req.Phone != "" {
		// 短信验证码登录
		if !verifySMSCode(db, req.Phone, req.SMSCode) {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "短信验证码错误或已过期",
			})
			return
		}

		// 根据手机号查找用户
		if err := db.Where("phone = ?", req.Phone).First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusBadRequest, gin.H{
					"code":    400,
					"message": "该手机号未注册",
				})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "查询用户失败",
				})
			}
			return
		}
	} else {
		// 密码登录
		if err := db.Where("username = ?", req.Username).First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusBadRequest, gin.H{
					"code":    400,
					"message": "用户名或密码错误",
				})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    500,
					"message": "查询用户失败",
				})
			}
			return
		}

		// 验证密码
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "用户名或密码错误",
			})
			return
		}
	}

	// 检查用户状态
	if user.Status != 1 {
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "账号已被禁用",
		})
		return
	}

	// 生成JWT令牌
	token, err := middleware.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "生成令牌失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"user":  user,
			"token": token,
		},
		"message": "登录成功",
	})
}

// SendSMSCode 发送短信验证码
func (h *UserHandler) SendSMSCode(c *gin.Context) {
	var req struct {
		Phone string `json:"phone" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误：" + err.Error(),
		})
		return
	}

	// 验证手机号格式
	if !utils.IsValidPhone(req.Phone) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "手机号格式不正确",
		})
		return
	}

	db := database.GetDB()

	// 生成验证码
	code := utils.GenerateSMSCode()
	expiresAt := time.Now().Add(5 * time.Minute) // 5分钟有效期

	// 保存验证码到数据库
	smsCode := model.SMSCode{
		Phone:     req.Phone,
		Code:      code,
		ExpiresAt: expiresAt,
		Used:      0,
	}

	if err := db.Create(&smsCode).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "发送失败：" + err.Error(),
		})
		return
	}

	// TODO: 实际项目中需要调用短信服务商API发送短信
	// 这里仅打印到日志
	// log.Printf("向 %s 发送验证码：%s", req.Phone, code)

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "验证码已发送 (测试环境请查看日志)",
		// 测试用：返回验证码 (生产环境请删除)
		"test_code": code,
	})
}

// VerifySMSCode 验证短信验证码
func (h *UserHandler) VerifySMSCode(c *gin.Context) {
	var req struct {
		Phone string `json:"phone" binding:"required"`
		Code  string `json:"code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误：" + err.Error(),
		})
		return
	}

	db := database.GetDB()

	if verifySMSCode(db, req.Phone, req.Code) {
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "验证成功",
		})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "验证码错误或已过期",
		})
	}
}

// verifySMSCode 内部函数：验证短信验证码
func verifySMSCode(db *gorm.DB, phone, code string) bool {
	var smsCode model.SMSCode
	
	// 查找未使用且未过期的验证码
	err := db.Where("phone = ? AND code = ? AND used = 0 AND expires_at > ?", 
		phone, code, time.Now()).First(&smsCode).Error
	
	if err != nil {
		return false
	}

	// 标记为已使用
	db.Model(&smsCode).Update("used", 1)

	return true
}

// GetProfile 获取当前用户信息
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": "未登录",
		})
		return
	}

	db := database.GetDB()
	var user model.User

	if err := db.First(&user, userID.(uint64)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "用户不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"data":    user,
		"message": "success",
	})
}

// GetUsers 获取用户列表 (管理员)
func (h *UserHandler) GetUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	db := database.GetDB()
	var users []model.User
	var total int64

	if err := db.Model(&model.User{}).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询失败",
		})
		return
	}

	offset := (page - 1) * pageSize
	if err := db.Limit(pageSize).Offset(offset).Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "查询失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"list":      users,
			"total":     total,
			"page":      page,
			"pageSize":  pageSize,
		},
		"message": "success",
	})
}

// UpdateUserRole 更新用户角色 (管理员)
func (h *UserHandler) UpdateUserRole(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的用户ID",
		})
		return
	}

	var req struct {
		Role int `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误",
		})
		return
	}

	// 角色验证
	if req.Role < 1 || req.Role > 3 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "无效的角色值",
		})
		return
	}

	db := database.GetDB()
	result := db.Model(&model.User{}).Where("id = ?", id).Update("role", req.Role)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "更新失败",
		})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "用户不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "更新成功",
	})
}
