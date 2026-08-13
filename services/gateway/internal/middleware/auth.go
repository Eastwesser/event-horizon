// services/gateway/internal/middleware/auth.go
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	authPb "github.com/Eastwesser/event-horizon/services/auth/proto"
)

const (
	CtxUserID = "user_id"
	CtxEmail  = "email"
	CtxRole   = "role"
)

// ExtractBearerToken pulls the raw JWT out of an "Authorization: Bearer <token>" header.
func ExtractBearerToken(header string) (string, bool) {
	if header == "" {
		return "", false
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return strings.TrimSpace(parts[1]), true
	}
	// Some clients send the raw token without the "Bearer " prefix.
	return strings.TrimSpace(header), true
}

// AuthClient matches client.AuthClient's gRPC surface needed here (avoids an import cycle).
type AuthClient interface {
	GetClient() authPb.AuthServiceClient
}

// RequireAuth validates the bearer token against the real Auth service (ValidateToken RPC,
// which itself checks the Redis session) and stores user_id/email/role in the gin context.
// This replaces the old local base64 JWT decode, which trusted the token's claims blindly
// and never checked for revocation (logout) or the source of truth in Postgres/Redis.
func RequireAuth(authClient AuthClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := ExtractBearerToken(c.GetHeader("Authorization"))
		if !ok || token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		resp, err := authClient.GetClient().ValidateToken(c.Request.Context(), &authPb.ValidateTokenRequest{Token: token})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "auth service unavailable"})
			return
		}
		if !resp.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set(CtxUserID, resp.UserId)
		c.Set(CtxEmail, resp.Email)
		c.Set(CtxRole, resp.Role)
		c.Next()
	}
}

// RequireRole must run after RequireAuth. It rejects the request unless the
// authenticated user's role is one of the allowed roles.
func RequireRole(allowed ...string) gin.HandlerFunc {
	allowedSet := make(map[string]bool, len(allowed))
	for _, r := range allowed {
		allowedSet[r] = true
	}
	return func(c *gin.Context) {
		role, _ := c.Get(CtxRole)
		roleStr, _ := role.(string)
		if !allowedSet[roleStr] {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			return
		}
		c.Next()
	}
}

// UserID is a small helper for handlers to read the authenticated user id
// set by RequireAuth, instead of re-parsing the token.
func UserID(c *gin.Context) string {
	v, _ := c.Get(CtxUserID)
	s, _ := v.(string)
	return s
}

// Role is a small helper for handlers to read the authenticated user's role.
func Role(c *gin.Context) string {
	v, _ := c.Get(CtxRole)
	s, _ := v.(string)
	return s
}
