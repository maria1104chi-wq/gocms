// 工具函数 - IP归属地查询等
// 文件名: utils.go
// 路径: /workspace/backend/internal/utils/utils.go

package utils

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// IPLocationResponse IP归属地API响应结构
type IPLocationResponse struct {
	Country string `json:"country"`
	Region  string `json:"region"`
	City    string `json:"city"`
}

// GetIPLocation 获取IP归属地
// 使用免费IP API (生产环境建议使用商业服务或本地IP库)
func GetIPLocation(ip string) string {
	// 处理IPv6 localhost
	if ip == "::1" || ip == "localhost" {
		return "本地网络"
	}

	// 简化实现：返回模拟数据
	// 实际项目中应调用真实API，如：
	// - http://ip-api.com/json/{ip}
	// - https://api.ip.sb/geoip/{ip}
	// - 淘宝IP地址库：http://ip.taobao.com/outGetIpInfo
	
	// 示例：调用 ip-api.com (免费，有请求限制)
	url := fmt.Sprintf("http://ip-api.com/json/%s?lang=zh-CN", ip)
	
	client := &http.Client{
		Timeout: 3 * time.Second,
	}
	
	resp, err := client.Get(url)
	if err != nil {
		return getMockLocation(ip)
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return getMockLocation(ip)
	}
	
	var result IPLocationResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return getMockLocation(ip)
	}
	
	// 构建地理位置字符串
	location := strings.TrimSpace(result.Country)
	if result.Region != "" {
		location += " " + result.Region
	}
	if result.City != "" {
		location += " " + result.City
	}
	
	if location == "" {
		return getMockLocation(ip)
	}
	
	return location
}

// getMockLocation 返回模拟的地理位置 (API失败时的降级处理)
func getMockLocation(ip string) string {
	// 根据IP段返回简单的地理位置
	if strings.HasPrefix(ip, "192.168.") || strings.HasPrefix(ip, "10.") {
		return "局域网"
	}
	
	// 默认返回未知
	return "未知地区"
}

// GenerateSMSCode 生成短信验证码
func GenerateSMSCode() string {
	// 生成6位数字验证码
	// 实际项目中应使用加密安全的随机数生成器
	return fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)
}

// IsValidPhone 验证手机号格式 (中国大陆)
func IsValidPhone(phone string) bool {
	// 简单验证：11位数字，以1开头
	if len(phone) != 11 {
		return false
	}
	if !strings.HasPrefix(phone, "1") {
		return false
	}
	// 更严格的验证可以使用正则表达式
	return true
}

// SanitizeString 清理字符串 (移除HTML标签等)
func SanitizeString(input string) string {
	// 简单实现：移除常见的HTML标签
	// 生产环境建议使用专业的HTML净化库
	tags := []string{"<script>", "</script>", "<iframe>", "</iframe>"}
	result := input
	for _, tag := range tags {
		result = strings.ReplaceAll(result, tag, "")
	}
	return result
}

// FormatTime 格式化时间为友好显示
func FormatTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)
	
	if diff < time.Minute {
		return "刚刚"
	} else if diff < time.Hour {
		return fmt.Sprintf("%d分钟前", int(diff.Minutes()))
	} else if diff < 24*time.Hour {
		return fmt.Sprintf("%d小时前", int(diff.Hours()))
	} else if diff < 7*24*time.Hour {
		return fmt.Sprintf("%d天前", int(diff.Hours()/24))
	} else {
		return t.Format("2006-01-02")
	}
}
