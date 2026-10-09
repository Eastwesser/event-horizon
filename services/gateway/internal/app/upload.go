package app

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Eastwesser/event-horizon/services/gateway/internal/client"
	"github.com/Eastwesser/event-horizon/services/gateway/internal/config"
	"github.com/Eastwesser/event-horizon/services/gateway/internal/middleware"
)

var allowedUploadExt = map[string]struct{}{
	".jpg":  {},
	".jpeg": {},
	".png":  {},
	".webp": {},
}

func ensureUploadDir(dir string) error {
	return os.MkdirAll(dir, 0o755)
}

// registerUploadRoutes serves local inventory images and accepts multipart uploads.
// Scoped to author/admin — kids-safe catalog images only, not a general file host.
func registerUploadRoutes(r *gin.Engine, cfg *config.Config, authClient *client.AuthClient) {
	if err := ensureUploadDir(cfg.UploadDir); err != nil {
		log.Printf("⚠️ upload dir %s: %v (uploads disabled)", cfg.UploadDir, err)
		return
	}
	log.Printf("📁 Uploads: dir=%s max=%d bytes", cfg.UploadDir, cfg.UploadMaxBytes)

	r.Static("/uploads", cfg.UploadDir)

	r.POST("/api/v1/uploads", middleware.RequireAuth(authClient), middleware.RequireRole(RoleAuthor, RoleAdmin), func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, cfg.UploadMaxBytes+1024)
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "file required (multipart field \"file\")"})
			return
		}
		if file.Size <= 0 || file.Size > cfg.UploadMaxBytes {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("file size must be 1..%d bytes", cfg.UploadMaxBytes),
			})
			return
		}

		ext := strings.ToLower(filepath.Ext(file.Filename))
		if _, ok := allowedUploadExt[ext]; !ok {
			c.JSON(http.StatusBadRequest, gin.H{"error": "only .jpg, .jpeg, .png, .webp allowed"})
			return
		}
		// Normalize jpeg → .jpg
		if ext == ".jpeg" {
			ext = ".jpg"
		}

		src, err := file.Open()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot read upload"})
			return
		}
		defer src.Close()

		head := make([]byte, 512)
		n, _ := io.ReadFull(src, head)
		detected := http.DetectContentType(head[:n])
		if !strings.HasPrefix(detected, "image/") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "content is not an image"})
			return
		}

		name := uuid.New().String() + ext
		destPath := filepath.Join(cfg.UploadDir, name)
		dst, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0o644)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot store upload"})
			return
		}
		defer dst.Close()

		if _, err := dst.Write(head[:n]); err != nil {
			_ = os.Remove(destPath)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "write failed"})
			return
		}
		if _, err := io.Copy(dst, src); err != nil {
			_ = os.Remove(destPath)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "write failed"})
			return
		}

		url := "/uploads/" + name
		c.JSON(http.StatusOK, gin.H{"url": url, "size": file.Size})
	})
}
