package handlers

import (
	"net/http"

	"github.com/wb-go/wbf/ginext"
	"github.com/wb_technoschool/level_3/task_2/models"
	"github.com/wb_technoschool/level_3/task_2/service"
	"github.com/wb_technoschool/level_3/task_2/storage"
)

type Handler struct {
	service *service.ShortenerService
}

func NewHandler(service *service.ShortenerService) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) Shorten(c *ginext.Context) {
	var req models.ShortenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	url, err := h.service.Shorten(c.Request.Context(), &req)
	if err != nil {
		if err == storage.ErrShortURLTaken {
			c.JSON(http.StatusConflict, ginext.H{"error": "short URL already taken"})
			return
		}
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, ginext.H{
		"short_url": url.ShortURL,
		"long_url":  url.LongURL,
	})
}

func (h *Handler) Redirect(c *ginext.Context) {
	shortURL := c.Param("short_url")

	longURL, err := h.service.Redirect(c.Request.Context(), shortURL)
	if err != nil {
		if err == storage.ErrURLNotFound {
			c.JSON(http.StatusNotFound, ginext.H{"error": "short URL not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	userAgent := c.GetHeader("User-Agent")
	ip := c.ClientIP()

	go func() {
		h.service.RecordClick(c.Request.Context(), shortURL, userAgent, ip)
	}()

	c.Redirect(http.StatusFound, longURL)
}

func (h *Handler) Analytics(c *ginext.Context) {
	shortURL := c.Param("short_url")

	analytics, err := h.service.GetAnalytics(c.Request.Context(), shortURL)
	if err != nil {
		if err == storage.ErrURLNotFound {
			c.JSON(http.StatusNotFound, ginext.H{"error": "short URL not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, analytics)
}
