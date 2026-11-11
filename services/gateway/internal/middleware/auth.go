// services/gateway/internal/middleware/auth.go
package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	authPb "github.com/Eastwesser/event-horizon/services/auth/proto"
)

const (
	CtxUserID = "user_id"
	CtxEmail  = "email"
	CtxRole   = "role"
)

// validateCacheTTL is the in-memory TTL for a validated token.
// 1s is a compromise: under load (140 VU × 4 calls ≈ 560 ValidateToken/s
// but only ~10 distinct tokens) it collapses the gRPC fan-out ~50×, while
// a logout still returns 401 within 1s. Read paths (shop/items, profile)
// tolerate that; write paths re-validate on the next tick.
//
// In-memory (not Redis): 10 keys × 3 fields × 1s is negligible, and
// avoiding a network hop matters more than sharing across replicas —
// each gateway replica keeps its own tiny cache.
const validateCacheTTL = 1 * time.Second

type cachedAuth struct {
	userID string
	email  string
	role   string
	expiry time.Time
}

var (
	validateCache   sync.Map // token → cachedAuth
	validateCacheMu sync.Mutex
)

func cacheGet(token string) (cachedAuth, bool) {
	v, ok := validateCache.Load(token)
	if !ok {
		return cachedAuth{}, false
	}
	c := v.(cachedAuth)
	if time.Now().After(c.expiry) {
		validateCache.Delete(token)
		return cachedAuth{}, false
	}
	return c, true
}

func cacheSet(token string, c cachedAuth) {
	c.expiry = time.Now().Add(validateCacheTTL)
	validateCache.Store(token, c)
}

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
//
// ValidateToken is cached in-memory for validateCacheTTL (1s). Without it, every
// closed request under load re-hits Auth (560 ValidateToken/s at P=200, ~10 distinct
// tokens → most of those calls are duplicate). 1s keeps logout latency acceptable
// (401 within 1s) while collapsing the fan-out ~50×.
func RequireAuth(authClient AuthClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := ExtractBearerToken(c.GetHeader("Authorization"))
		if !ok || token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		// Fast path: in-memory cache hit within TTL.
		if cached, hit := cacheGet(token); hit {
			c.Set(CtxUserID, cached.userID)
			c.Set(CtxEmail, cached.email)
			c.Set(CtxRole, cached.role)
			c.Next()
			return
		}

		// Slow path: real ValidateToken RPC.
		resp, err := authClient.GetClient().ValidateToken(c.Request.Context(), &authPb.ValidateTokenRequest{Token: token})
		if err != nil {
			if st, ok := status.FromError(err); ok && st.Code() == codes.Unavailable {
				c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "session store unavailable"})
				return
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "auth service unavailable"})
			return
		}
		if !resp.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		// Cache only valid results. A revoked token (logout) will still be
		// honoured for up to 1s — acceptable for read paths.
		cacheSet(token, cachedAuth{
			userID: resp.UserId,
			email:  resp.Email,
			role:   resp.Role,
		})

		c.Set(CtxUserID, resp.UserId)
		c.Set(CtxEmail, resp.Email)
		c.Set(CtxRole, resp.Role)
		c.Next()
	}
}

// // RequireAuth validates the bearer token against the real Auth service (ValidateToken RPC,
// // which itself checks the Redis session) and stores user_id/email/role in the gin context.
// // This replaces the old local base64 JWT decode, which trusted the token's claims blindly
// // and never checked for revocation (logout) or the source of truth in Postgres/Redis.
// func RequireAuth(authClient AuthClient) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		token, ok := ExtractBearerToken(c.GetHeader("Authorization"))
// 		if !ok || token == "" {
// 			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
// 			return
// 		}

// 		resp, err := authClient.GetClient().ValidateToken(c.Request.Context(), &authPb.ValidateTokenRequest{Token: token})
// 		if err != nil {
// 			if st, ok := status.FromError(err); ok && st.Code() == codes.Unavailable {
// 				c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "session store unavailable"})
// 				return
// 			}
// 			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "auth service unavailable"})
// 			return
// 		}
// 		if !resp.Valid {
// 			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
// 			return
// 		}

// 		c.Set(CtxUserID, resp.UserId)
// 		c.Set(CtxEmail, resp.Email)
// 		c.Set(CtxRole, resp.Role)
// 		c.Next()
// 	}
// }

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

// Email is a small helper for handlers to read the authenticated user's email.
func Email(c *gin.Context) string {
	v, _ := c.Get(CtxEmail)
	s, _ := v.(string)
	return s
}
