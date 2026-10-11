package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const maxJSONBody = 1 << 20 // 1 MiB. Uploads use their own limit.

// SecurityHeaders sets the baseline headers that do not depend on cookies.
// The API is Bearer-token, so a CSRF token is not added here.
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
		h.Set("X-Permitted-Cross-Domain-Policies", "none")
		h.Set("Content-Security-Policy", contentSecurityPolicy(c.Request.URL.Path))
		if c.Request.TLS != nil || strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https") {
			h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		c.Next()
	}
}

func contentSecurityPolicy(path string) string {
	if path == "/docs" || strings.HasPrefix(path, "/docs/") {
		return "default-src 'none'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self'; frame-ancestors 'none'; base-uri 'none'"
	}
	return "default-src 'none'; frame-ancestors 'none'; base-uri 'none'; form-action 'none'"
}

// RejectUnsafeRequest drops path traversal and methods this API does not use.
func RejectUnsafeRequest() gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := c.Request.URL.RawPath
		if raw == "" {
			raw = c.Request.URL.Path
		}
		if strings.Contains(raw, "..") || strings.ContainsRune(raw, 0) || strings.Contains(raw, "\\") {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		switch c.Request.Method {
		case http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions:
		default:
			c.AbortWithStatus(http.StatusMethodNotAllowed)
			return
		}
		c.Next()
	}
}

// LimitJSONBody caps ordinary request bodies. Upload routes keep UploadMaxBytes.
func LimitJSONBody() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body == nil || strings.Contains(c.Request.URL.Path, "/uploads") {
			c.Next()
			return
		}
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxJSONBody)
		c.Next()
	}
}
