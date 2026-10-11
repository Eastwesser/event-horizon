package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestDocsGate_HiddenWithoutPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/docs", DocsGate("docs", ""), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/docs", nil))
	if w.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", w.Code)
	}
}

func TestDocsGate_RejectsWrongPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/docs", DocsGate("docs", "s3cret"), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/docs", nil)
	req.SetBasicAuth("docs", "nope")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("want 401, got %d", w.Code)
	}
	if w.Header().Get("WWW-Authenticate") == "" {
		t.Fatal("missing WWW-Authenticate")
	}
}

func TestDocsGate_AllowsPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/docs", DocsGate("docs", "s3cret"), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/docs", nil)
	req.SetBasicAuth("docs", "s3cret")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d", w.Code)
	}
}

func TestRejectUnsafeRequest_Traversal(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(RejectUnsafeRequest())
	r.GET("/files/*rest", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/files/ok", nil)
	req.URL.RawPath = "/files/../secret"
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", w.Code)
	}
}

func TestSecurityHeaders_CSP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(SecurityHeaders())
	r.GET("/health", func(c *gin.Context) { c.Status(http.StatusOK) })
	r.GET("/docs", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/health", nil))
	if !strings.Contains(w.Header().Get("Content-Security-Policy"), "default-src 'none'") {
		t.Fatalf("api csp: %s", w.Header().Get("Content-Security-Policy"))
	}

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/docs", nil))
	if !strings.Contains(w.Header().Get("Content-Security-Policy"), "script-src 'self'") {
		t.Fatalf("docs csp: %s", w.Header().Get("Content-Security-Policy"))
	}
}
