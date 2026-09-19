package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type visitor struct {
	count int
	reset time.Time
}
type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]visitor
	limit    int
	window   time.Duration
}

func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{visitors: make(map[string]visitor), limit: limit, window: window}
}
func (r *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP()
		now := time.Now()
		r.mu.Lock()
		v := r.visitors[key]
		if v.reset.Before(now) {
			v = visitor{count: 0, reset: now.Add(r.window)}
		}
		v.count++
		r.visitors[key] = v
		if v.count > r.limit {
			r.mu.Unlock()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too many requests"})
			return
		}
		r.mu.Unlock()
		c.Next()
	}
}
