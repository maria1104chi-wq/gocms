// Redis连接初始化
// 文件名: redis.go
// 路径: /workspace/backend/internal/database/redis.go

package database

import (
	"context"
	"fmt"
	"log"

	"blog-cms/internal/config"

	"github.com/go-redis/redis/v8"
)

var RedisClient *redis.Client
var ctx = context.Background()

// InitRedis 初始化Redis连接
func InitRedis(cfg *config.RedisConfig) error {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     cfg.Host + ":" + cfg.Port,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// 测试连接
	_, err := RedisClient.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("连接Redis失败：%v", err)
	}

	log.Println("Redis连接成功")
	return nil
}

// GetRedis 获取Redis客户端实例
func GetRedis() *redis.Client {
	return RedisClient
}
