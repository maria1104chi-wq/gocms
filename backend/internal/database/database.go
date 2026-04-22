// 数据库连接初始化
// 文件名: database.go
// 路径: /workspace/backend/internal/database/database.go

package database

import (
	"fmt"
	"log"

	"blog-cms/internal/config"
	"blog-cms/internal/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDB 初始化数据库连接
func InitDB(cfg *config.DatabaseConfig) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
	)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		return fmt.Errorf("连接数据库失败：%v", err)
	}

	// 自动迁移表结构
	err = autoMigrate()
	if err != nil {
		return fmt.Errorf("数据库迁移失败：%v", err)
	}

	log.Println("数据库连接成功并完成迁移")
	return nil
}

// autoMigrate 自动迁移所有模型表结构
func autoMigrate() error {
	return DB.AutoMigrate(
		&model.User{},
		&model.Category{},
		&model.UserCategory{},
		&model.Article{},
		&model.Comment{},
		&model.SensitiveWord{},
		&model.SMSCode{},
		&model.VisitLog{},
		&model.Like{},
		&model.SystemConfig{},
	)
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}
