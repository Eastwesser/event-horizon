package handlers

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb_technoschool/level_3/task_4/models"
	"github.com/wb_technoschool/level_3/task_4/queue"
	"github.com/wb_technoschool/level_3/task_4/storage"
)

type Handler struct {
	storage   storage.Storage
	fileStore *storage.FileStorage
	queue     queue.Queue
}

func NewHandler(
	storage storage.Storage,
	fileStore *storage.FileStorage,
	queue queue.Queue,
) *Handler {
	return &Handler{
		storage:   storage,
		fileStore: fileStore,
		queue:     queue,
	}
}

func (h *Handler) Upload(c *ginext.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": "file is required"})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}
	defer src.Close()

	imageID := uuid.New().String()
	data, err := io.ReadAll(src)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	ext := filepath.Ext(file.Filename)
	if ext == "" {
		ext = ".jpg"
	}

	originalPath, err := h.fileStore.SaveFile(imageID, data, "original", ext)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	now := time.Now()
	image := &models.Image{
		ID:           imageID,
		OriginalPath: originalPath,
		Status:       models.StatusPending,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := h.storage.Save(image); err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	task := &models.ProcessingTask{
		ImageID: imageID,
		Action:  "process",
	}

	if err := h.queue.Publish(c.Request.Context(), task); err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, ginext.H{
		"id":     imageID,
		"status": image.Status,
	})
}

func (h *Handler) GetImage(c *ginext.Context) {
	id := c.Param("id")

	image, err := h.storage.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, ginext.H{"error": "image not found"})
		return
	}

	c.JSON(http.StatusOK, image)
}

func (h *Handler) GetImageFile(c *ginext.Context) {
	id := c.Param("id")
	imageType := c.DefaultQuery("type", "processed")

	image, err := h.storage.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, ginext.H{"error": "image not found"})
		return
	}

	var filePath string
	switch imageType {
	case "original":
		filePath = image.OriginalPath
	case "thumbnail":
		filePath = image.ThumbnailPath
	default:
		filePath = image.ProcessedPath
	}

	if filePath == "" || image.Status != models.StatusCompleted {
		c.JSON(http.StatusNotFound, ginext.H{"error": "image not processed yet"})
		return
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, ginext.H{"error": "file not found"})
		return
	}

	c.File(filePath)
}

func (h *Handler) DeleteImage(c *ginext.Context) {
	id := c.Param("id")

	image, err := h.storage.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, ginext.H{"error": "image not found"})
		return
	}

	if image.OriginalPath != "" {
		os.Remove(image.OriginalPath)
	}
	if image.ProcessedPath != "" {
		os.Remove(image.ProcessedPath)
	}
	if image.ThumbnailPath != "" {
		os.Remove(image.ThumbnailPath)
	}

	if err := h.storage.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ginext.H{"message": "image deleted"})
}

func (h *Handler) GetAllImages(c *ginext.Context) {
	images, err := h.storage.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, images)
}

