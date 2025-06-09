package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type IPRateLimiter struct {
	ips    map[string]*rateLimiterWithTime
	mu     *sync.RWMutex
	rate   rate.Limit
	burst  int
	expiry time.Duration
}

type rateLimiterWithTime struct {
	limiter   *rate.Limiter
	lastUsage time.Time
}

func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	i := &IPRateLimiter{
		ips:    make(map[string]*rateLimiterWithTime),
		mu:     &sync.RWMutex{},
		rate:   r,
		burst:  b,
		expiry: time.Hour, // Clean up unused limiters after 1 hour
	}

	// Start cleanup routine
	go i.cleanupRoutine()
	return i
}

func (i *IPRateLimiter) cleanupRoutine() {
	ticker := time.NewTicker(time.Minute)
	for range ticker.C {
		i.mu.Lock()
		for ip, wrapper := range i.ips {
			if time.Since(wrapper.lastUsage) > i.expiry {
				delete(i.ips, ip)
			}
		}
		i.mu.Unlock()
	}
}

func (i *IPRateLimiter) getLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	wrapper, exists := i.ips[ip]
	if !exists {
		wrapper = &rateLimiterWithTime{
			limiter:   rate.NewLimiter(i.rate, i.burst),
			lastUsage: time.Now(),
		}
		i.ips[ip] = wrapper
	} else {
		wrapper.lastUsage = time.Now()
	}

	return wrapper.limiter
}

func RateLimit(rps float64, burst int) gin.HandlerFunc {
	limiter := NewIPRateLimiter(rate.Limit(rps), burst)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !limiter.getLimiter(ip).Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "too many requests",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
