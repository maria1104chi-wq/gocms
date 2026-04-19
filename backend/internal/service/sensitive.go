// 敏感词服务
// 文件名: sensitive.go
// 路径: /workspace/backend/internal/service/sensitive.go

package service

import (
	"context"
	"strings"
	"sync"

	"blog-cms/internal/database"
	"blog-cms/internal/model"
)

// SensitiveService 敏感词服务
type SensitiveService struct {
	mu      sync.RWMutex
	words   map[string]bool // 使用map存储敏感词，便于快速查找
	loaded  bool
}

var sensitiveService *SensitiveService

// GetSensitiveService 获取敏感词服务单例
func GetSensitiveService() *SensitiveService {
	if sensitiveService == nil {
		sensitiveService = &SensitiveService{
			words:  make(map[string]bool),
			loaded: false,
		}
	}
	return sensitiveService
}

// LoadWords 从数据库加载敏感词到内存
func (s *SensitiveService) LoadWords() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	db := database.GetDB()
	var words []model.SensitiveWord

	if err := db.Find(&words).Error; err != nil {
		return err
	}

	// 清空旧数据
	s.words = make(map[string]bool)

	// 加载新数据
	for _, word := range words {
		s.words[word.Word] = true
	}

	s.loaded = true
	return nil
}

// Filter 过滤文本中的敏感词，替换为***
func (s *SensitiveService) Filter(text string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.loaded {
		return text
	}

	result := text
	for word := range s.words {
		if strings.Contains(result, word) {
			// 替换敏感词为***
			result = strings.ReplaceAll(result, word, "***")
		}
	}

	return result
}

// Contains 检查文本是否包含敏感词
func (s *SensitiveService) Contains(text string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.loaded {
		return false
	}

	for word := range s.words {
		if strings.Contains(text, word) {
			return true
		}
	}

	return false
}

// AddWord 添加敏感词到数据库并更新内存
func (s *SensitiveService) AddWord(word, category string) error {
	db := database.GetDB()

	// 先插入数据库
	sensitiveWord := model.SensitiveWord{
		Word:     word,
		Category: category,
	}

	if err := db.Create(&sensitiveWord).Error; err != nil {
		return err
	}

	// 更新内存
	return s.LoadWords()
}

// DeleteWord 从数据库删除敏感词并更新内存
func (s *SensitiveService) DeleteWord(word string) error {
	db := database.GetDB()

	if err := db.Where("word = ?", word).Delete(&model.SensitiveWord{}).Error; err != nil {
		return err
	}

	// 更新内存
	return s.LoadWords()
}

// SyncWithRedis 将敏感词同步到Redis缓存 (用于分布式部署)
func (s *SensitiveService) SyncWithRedis(ctx context.Context) error {
	redisClient := database.GetRedis()
	
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 将所有敏感词存入Redis Set
	words := make([]interface{}, 0, len(s.words))
	for word := range s.words {
		words = append(words, word)
	}

	if len(words) > 0 {
		// 删除旧的集合
		redisClient.Del(ctx, "sensitive:words")
		// 创建新的集合
		redisClient.SAdd(ctx, "sensitive:words", words...)
	}

	return nil
}

// LoadFromRedis 从Redis加载敏感词到内存
func (s *SensitiveService) LoadFromRedis(ctx context.Context) error {
	redisClient := database.GetRedis()

	// 从Redis获取所有敏感词
	words, err := redisClient.SMembers(ctx, "sensitive:words").Result()
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.words = make(map[string]bool)
	for _, word := range words {
		s.words[word] = true
	}

	s.loaded = true
	return nil
}
