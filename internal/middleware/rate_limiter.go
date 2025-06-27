package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
	ginlimiter "github.com/ulule/limiter/v3/drivers/middleware/gin"
	memory "github.com/ulule/limiter/v3/drivers/store/memory"
)

// RateLimiterMiddleware returns a Gin middleware that limits requests per IP.
func RateLimiterMiddleware() gin.HandlerFunc {
	// 100 requests per minute per IP
	rate, _ := limiter.NewRateFromFormatted("100-M")
	store := memory.NewStore()
	instance := limiter.New(store, rate)
	return ginlimiter.NewMiddleware(instance)
}
