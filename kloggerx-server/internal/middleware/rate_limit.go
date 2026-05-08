package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter 令牌桶限流器
type RateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*tokenBucket
	rate     float64       // 每秒生成的令牌数
	capacity int           // 桶容量
	cleanup  time.Duration // 清理间隔
}

type tokenBucket struct {
	tokens   float64
	lastTime time.Time
}

// NewRateLimiter 创建限流器
// rate: 每秒允许的请求数
// capacity: 突发容量(桶大小)
func NewRateLimiter(rate float64, capacity int) *RateLimiter {
	rl := &RateLimiter{
		buckets:  make(map[string]*tokenBucket),
		rate:     rate,
		capacity: capacity,
		cleanup:  5 * time.Minute,
	}
	go rl.cleanupLoop()
	return rl
}

// 定期清理过期桶
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cleanup)
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, bucket := range rl.buckets {
			if now.Sub(bucket.lastTime) > 10*time.Minute {
				delete(rl.buckets, key)
			}
		}
		rl.mu.Unlock()
	}
}

// Allow 检查是否允许请求
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	bucket, ok := rl.buckets[key]
	if !ok {
		bucket = &tokenBucket{tokens: float64(rl.capacity), lastTime: time.Now()}
		rl.buckets[key] = bucket
	}

	// 补充令牌
	now := time.Now()
	elapsed := now.Sub(bucket.lastTime).Seconds()
	bucket.tokens += elapsed * rl.rate
	if bucket.tokens > float64(rl.capacity) {
		bucket.tokens = float64(rl.capacity)
	}
	bucket.lastTime = now

	// 消费令牌
	if bucket.tokens >= 1 {
		bucket.tokens--
		return true
	}
	return false
}

// RateLimit 中间件 - 按用户ID限流
func RateLimit(rate float64, capacity int) gin.HandlerFunc {
	limiter := NewRateLimiter(rate, capacity)
	return func(c *gin.Context) {
		// 优先按用户ID限流，未登录则按IP
		key := c.ClientIP()
		if userId, exists := c.Get("userId"); exists {
			key = fmt.Sprintf("user:%v", userId)
		}

		if !limiter.Allow(key) {
			c.Header("Retry-After", "1")
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code":    429,
				"message": "请求过于频繁，请稍后重试",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// AIRateLimit AI接口专用限流(更严格)
func AIRateLimit() gin.HandlerFunc {
	// AI接口: 每秒2个请求，突发5个
	return RateLimit(2, 5)
}

// HeavyTaskLimit 重型任务全局并发限制
func HeavyTaskLimit(maxConcurrent int) gin.HandlerFunc {
	semaphore := make(chan struct{}, maxConcurrent)
	return func(c *gin.Context) {
		select {
		case semaphore <- struct{}{}:
			defer func() { <-semaphore }()
			c.Next()
		default:
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"code":    503,
				"message": "系统繁忙，请稍后重试",
			})
			c.Abort()
		}
	}
}
