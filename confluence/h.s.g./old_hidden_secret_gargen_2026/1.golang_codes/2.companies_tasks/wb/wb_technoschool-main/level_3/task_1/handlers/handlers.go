package handlers

import (
	"net/http"

	"github.com/wb-go/wbf/ginext"
	"github.com/wb_technoschool/level_3/task_1/models"
	"github.com/wb_technoschool/level_3/task_1/service"
	"github.com/wb_technoschool/level_3/task_1/storage"
)

type Handler struct {
	service *service.NotificationService
}

func NewHandler(service *service.NotificationService) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) CreateNotification(c *ginext.Context) {
	var req models.CreateNotificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	notification, err := h.service.CreateNotification(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, notification)
}

func (h *Handler) GetNotification(c *ginext.Context) {
	id := c.Param("id")

	notification, err := h.service.GetNotification(c.Request.Context(), id)
	if err != nil {
		if err == storage.ErrNotificationNotFound {
			c.JSON(http.StatusNotFound, ginext.H{"error": "notification not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, notification)
}

func (h *Handler) DeleteNotification(c *ginext.Context) {
	id := c.Param("id")

	err := h.service.DeleteNotification(c.Request.Context(), id)
	if err != nil {
		if err == storage.ErrNotificationNotFound {
			c.JSON(http.StatusNotFound, ginext.H{"error": "notification not found"})
			return
		}
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ginext.H{"message": "notification cancelled"})
}

func (h *Handler) GetAllNotifications(c *ginext.Context) {
	notifications, err := h.service.GetAllNotifications(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, notifications)
}
