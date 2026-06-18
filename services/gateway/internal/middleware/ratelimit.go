// services/gateway/internal/middleware/ratelimit.go
package middleware

import (
	"encoding/base64" 
	"encoding/json"
	"fmt" 
    "net/http"
    "strings"

    "github.com/gin-gonic/gin"

    "event_horizon/services/gateway/internal/ratelimit"
)

func RateLimitMiddleware(limiter *ratelimit.RateLimiter) gin.HandlerFunc {
    return func(c *gin.Context) {
        path := c.Request.URL.Path
        method := c.Request.Method

        // Получаем user_id из заголовка Authorization (если есть)
        userID := ""
        authHeader := c.GetHeader("Authorization")
        if authHeader != "" {
            // Пробуем извлечь user_id из токена (через getUserIDFromToken)
            parts := strings.Split(authHeader, " ")
            if len(parts) == 2 && parts[0] == "Bearer" {
                token := parts[1]
                // Используем функцию из main.go (нужно будет вынести или продублировать)
                if uid, err := getUserIDFromToken(token); err == nil {
                    userID = uid
                }
            }
        }

        var allowed bool
        // RATE LIMITER (OFF)
        // switch {
        // case path == "/api/game/submit" && method == "POST":
        //     // Временно отключаем для теста
        //     allowed = true  // 👈 вместо limiter.AllowSubmit()
        //     if userID != "" {
        //         allowed = true  // 👈 вместо limiter.AllowSubmit()
        //     }    
        //     // } else {
        //     //     allowed = limiter.AllowSubmit(c.ClientIP())
        //     // }

        // THE RATE LIMITER (ON)
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
            c.JSON(http.StatusTooManyRequests, gin.H{
                "error":       "Too many requests. Please try again later.",
                "retry_after": 1,
            })
            c.Abort()
            return
        }

        c.Next()
    }
}

// Вспомогательная функция для извлечения user_id из токена
func getUserIDFromToken(tokenString string) (string, error) {
    parts := strings.Split(tokenString, ".")
    if len(parts) != 3 {
        return "", fmt.Errorf("invalid token format")
    }

    payload, err := base64.RawURLEncoding.DecodeString(parts[1])
    if err != nil {
        return "", err
    }

    var claims map[string]interface{}
    if err := json.Unmarshal(payload, &claims); err != nil {
        return "", err
    }

    userID, ok := claims["user_id"].(string)
    if !ok {
        return "", fmt.Errorf("user_id not found in token")
    }

    return userID, nil
}
