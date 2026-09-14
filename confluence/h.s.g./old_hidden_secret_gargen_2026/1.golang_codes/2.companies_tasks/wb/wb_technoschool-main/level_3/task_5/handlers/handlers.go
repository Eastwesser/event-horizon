package handlers

import (
	"net/http"

	"github.com/wb-go/wbf/ginext"
	"github.com/wb_technoschool/level_3/task_5/models"
	"github.com/wb_technoschool/level_3/task_5/service"
)

type Handler struct {
	service *service.BookingService
}

func NewHandler(service *service.BookingService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) CreateEvent(c *ginext.Context) {
	var event models.Event
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	if err := h.service.CreateEvent(c.Request.Context(), &event); err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, event)
}

func (h *Handler) GetEvent(c *ginext.Context) {
	id := c.Param("id")

	event, err := h.service.GetEvent(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, ginext.H{"error": "event not found"})
		return
	}

	c.JSON(http.StatusOK, event)
}

func (h *Handler) GetAllEvents(c *ginext.Context) {
	events, err := h.service.GetAllEvents(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, events)
}

func (h *Handler) CreateBooking(c *ginext.Context) {
	eventID := c.Param("id")

	var req struct {
		UserEmail string `json:"user_email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": "valid email is required"})
		return
	}

	booking, err := h.service.CreateBooking(c.Request.Context(), eventID, req.UserEmail)
	if err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, booking)
}

func (h *Handler) ConfirmBooking(c *ginext.Context) {
	bookingID := c.Param("id")

	if err := h.service.ConfirmBooking(c.Request.Context(), bookingID); err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ginext.H{"message": "booking confirmed"})
}

func (h *Handler) GetEventBookings(c *ginext.Context) {
	eventID := c.Param("id")

	bookings, err := h.service.GetEventBookings(c.Request.Context(), eventID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, bookings)
}

func (h *Handler) GetUserBookings(c *ginext.Context) {
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, ginext.H{"error": "email is required"})
		return
	}

	bookings, err := h.service.GetUserBookings(c.Request.Context(), email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, bookings)
}
