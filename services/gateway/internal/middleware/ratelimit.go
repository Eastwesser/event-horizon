// services/gateway/internal/middleware/ratelimit.go
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Eastwesser/event-horizon/services/gateway/internal/ratelimit"
)

// RateLimitMiddleware enforces per-route rate limits using Redis-backed sliding windows.
// Authenticated requests are throttled per user_id (set by RequireAuth); everything else
// falls back to the client IP.
func RateLimitMiddleware(limiter *ratelimit.RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		method := c.Request.Method

		userID := UserID(c)

		var allowed bool

		switch {
		case path == "/api/game/submit" && method == "POST":
			if userID != "" {
				allowed = limiter.AllowSubmit(userID)
			} else {
				allowed = limiter.AllowSubmit(c.ClientIP())
			}

		case path == "/api/auth/login" && method == "POST":
			allowed = limiter.AllowLogin(c.ClientIP())

		case path == "/ws/leaderboard":
			allowed = limiter.AllowWebSocket(c.ClientIP())

		default:
			allowed = true
		}

		if !allowed {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":       "Too many requests. Please try again later.",
				"retry_after": 1,
			})
			return
		}

		c.Next()
	}
}
