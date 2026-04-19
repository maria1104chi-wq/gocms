// 配置文件
// 文件名: config.go
// 路径: /workspace/backend/internal/config/config.go

package config

import (
	"os"
)

// Config 配置结构体
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	SMS      SMSConfig
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port string
	Mode string // debug, release, test
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret     string
	ExpireHour int
}

// SMSConfig 短信配置
type SMSConfig struct {
	Provider string // aliyun, tencent
	AccessID string
	Secret   string
	SignName string // 短信签名
}

// LoadConfig 加载配置 (从环境变量读取，便于Docker部署)
func LoadConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Mode: getEnv("SERVER_MODE", "release"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "mysql"),
			Port:     getEnv("DB_PORT", "3306"),
			User:     getEnv("DB_USER", "root"),
			Password: getEnv("DB_PASSWORD", "rootpassword"),
			DBName:   getEnv("DB_NAME", "blog_cms"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "redis"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       0,
		},
		JWT: JWTConfig{
			Secret:     getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
			ExpireHour: 24,
		},
		SMS: SMSConfig{
			Provider: getEnv("SMS_PROVIDER", "aliyun"),
			AccessID: getEnv("SMS_ACCESS_ID", ""),
			Secret:   getEnv("SMS_SECRET", ""),
			SignName: getEnv("SMS_SIGN_NAME", "博客平台"),
		},
	}
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
