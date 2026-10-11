package middleware

import (
	"crypto/subtle"
	"net/http"

	"github.com/gin-gonic/gin"
)

// DocsGate hides OpenAPI and Swagger.
// An empty password answers 404, so a public host without DOCS_PASSWORD does not serve the spec.
// A configured password uses HTTP Basic. The browser prompts; the secret is not put in the URL.
func DocsGate(user, password string) gin.HandlerFunc {
	userB := []byte(user)
	passB := []byte(password)
	return func(c *gin.Context) {
		if len(passB) == 0 {
			c.AbortWithStatus(http.StatusNotFound)
			return
		}
		gotUser, gotPass, ok := c.Request.BasicAuth()
		match := subtle.ConstantTimeCompare([]byte(gotUser), userB) &
			subtle.ConstantTimeCompare([]byte(gotPass), passB)
		if !ok || match != 1 {
			c.Header("WWW-Authenticate", `Basic realm="Event Horizon docs", charset="UTF-8"`)
			c.Header("Cache-Control", "no-store")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Header("Cache-Control", "private, no-store")
		c.Next()
	}
}
