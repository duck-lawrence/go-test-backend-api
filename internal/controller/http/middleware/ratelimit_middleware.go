package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type visitor struct {
	Limiter  *rate.Limiter
	LastSeen time.Time
}

var (
	visitors = make(map[string]*visitor)
	mu       sync.Mutex
)

// get limiter
func getLimiter(key string, r rate.Limit, b int) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	v, exists := visitors[key]
	if !exists {
		limiter := rate.NewLimiter(r, b)
		visitors[key] = &visitor{Limiter: limiter, LastSeen: time.Now()}
		return limiter
	}

	v.LastSeen = time.Now()
	return v.Limiter
}

func RateLimitMiddleware(r rate.Limit, b int) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP()

		limiter := getLimiter(key, r, b)
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests",
			})
			return
		}
		c.Next()
	}
}

func StartCleanupJob(expireAfter time.Duration, interval time.Duration) {
	go func() {
		for {
			time.Sleep(interval)
			mu.Lock()
			for key, v := range visitors {
				if time.Since(v.LastSeen) > expireAfter {
					delete(visitors, key)
				}
			}
			mu.Unlock()
		}
	}()
}
