// services/gateway/internal/middleware/ratelimit.go
package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Eastwesser/event-horizon/services/gateway/internal/ratelimit"
)

// RateLimitMiddleware enforces per-route and global (~100 req/s) rate limits.
// Subject is user_id when RequireAuth already ran; otherwise a hash of the Bearer
// token (same session → same bucket) or the client IP.
func RateLimitMiddleware(limiter *ratelimit.RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		method := c.Request.Method
		subject := rateLimitSubject(c)

		var allowed bool

		switch {
		case path == "/api/v1/game/submit" && method == "POST":
			allowed = limiter.AllowSubmit(subject)

		case path == "/api/v1/auth/login" && method == "POST":
			allowed = limiter.AllowLogin(c.ClientIP())

		case path == "/ws/leaderboard":
			allowed = limiter.AllowWebSocket(c.ClientIP())

		case path == "/health" || path == "/ready" || path == "/metrics":
			allowed = true

		default:
			allowed = limiter.AllowGlobal(subject)
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

func rateLimitSubject(c *gin.Context) string {
	if uid := UserID(c); uid != "" {
		return "u:" + uid
	}
	if tok, ok := ExtractBearerToken(c.GetHeader("Authorization")); ok && tok != "" {
		sum := sha256.Sum256([]byte(tok))
		return "t:" + hex.EncodeToString(sum[:8])
	}
	return "ip:" + c.ClientIP()
}
