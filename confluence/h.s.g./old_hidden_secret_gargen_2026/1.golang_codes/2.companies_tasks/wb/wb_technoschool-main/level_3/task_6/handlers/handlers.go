package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/wb-go/wb_technoschool/level_3/task_6/models"
	"github.com/wb-go/wb_technoschool/level_3/task_6/service"
	"github.com/wb-go/wb_technoschool/level_3/task_6/storage"
)

type Handlers struct {
	service *service.Service
}

func New(svc *service.Service) *Handlers {
	return &Handlers{service: svc}
}

// CreateTransaction POST /items
func (h *Handlers) CreateTransaction(c *gin.Context) {
	var req struct {
		Type     string  `json:"type" binding:"required"`
		Amount   float64 `json:"amount" binding:"required"`
		Category string  `json:"category" binding:"required"`
		Date     string  `json:"date" binding:"required"`
		Comment  string  `json:"comment"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	parsedDate, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format, use YYYY-MM-DD"})
		return
	}

	tx := &models.Transaction{
		Type:     models.TransactionType(req.Type),
		Amount:   req.Amount,
		Category: req.Category,
		Date:     parsedDate,
		Comment:  req.Comment,
	}

	id, err := h.service.CreateTransaction(c.Request.Context(), tx)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

// GetTransaction GET /items/:id
func (h *Handlers) GetTransaction(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	tx, err := h.service.GetTransaction(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tx)
}

// ListTransactions GET /items
func (h *Handlers) ListTransactions(c *gin.Context) {
	filters := storage.ListFilters{}

	if fromStr := c.Query("from"); fromStr != "" {
		from, err := time.Parse("2006-01-02", fromStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid from date format"})
			return
		}
		filters.FromDate = &from
	}

	if toStr := c.Query("to"); toStr != "" {
		to, err := time.Parse("2006-01-02", toStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid to date format"})
			return
		}
		// Make inclusive by adding 1 day
		to = to.AddDate(0, 0, 1)
		filters.ToDate = &to
	}

	filters.Category = c.Query("category")
	filters.Type = c.Query("type")
	filters.SortBy = c.DefaultQuery("sort", "date")
	filters.SortOrder = c.DefaultQuery("order", "DESC")

	if limit := c.Query("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			filters.Limit = l
		}
	}

	if offset := c.Query("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil {
			filters.Offset = o
		}
	}

	transactions, err := h.service.ListTransactions(c.Request.Context(), filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if transactions == nil {
		transactions = []models.Transaction{}
	}

	c.JSON(http.StatusOK, transactions)
}

// UpdateTransaction PUT /items/:id
func (h *Handlers) UpdateTransaction(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req struct {
		Type     string  `json:"type" binding:"required"`
		Amount   float64 `json:"amount" binding:"required"`
		Category string  `json:"category" binding:"required"`
		Date     string  `json:"date" binding:"required"`
		Comment  string  `json:"comment"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	parsedDate, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format, use YYYY-MM-DD"})
		return
	}

	tx := &models.Transaction{
		Type:     models.TransactionType(req.Type),
		Amount:   req.Amount,
		Category: req.Category,
		Date:     parsedDate,
		Comment:  req.Comment,
	}

	if err := h.service.UpdateTransaction(c.Request.Context(), id, tx); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

// DeleteTransaction DELETE /items/:id
func (h *Handlers) DeleteTransaction(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.service.DeleteTransaction(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// GetAnalytics GET /analytics
func (h *Handlers) GetAnalytics(c *gin.Context) {
	fromStr := c.Query("from")
	toStr := c.Query("to")

	if fromStr == "" || toStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "from and to dates are required"})
		return
	}

	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid from date format"})
		return
	}

	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid to date format"})
		return
	}

	// Make inclusive
	to = to.AddDate(0, 0, 1)

	analytics, err := h.service.GetAnalytics(c.Request.Context(), from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, analytics)
}

// GetCategoryAnalytics GET /analytics/categories
func (h *Handlers) GetCategoryAnalytics(c *gin.Context) {
	fromStr := c.Query("from")
	toStr := c.Query("to")

	if fromStr == "" || toStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "from and to dates are required"})
		return
	}

	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid from date format"})
		return
	}

	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid to date format"})
		return
	}

	// Make inclusive
	to = to.AddDate(0, 0, 1)

	analytics, err := h.service.GetCategoryAnalytics(c.Request.Context(), from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if analytics == nil {
		analytics = []models.CategoryAnalytics{}
	}

	c.JSON(http.StatusOK, analytics)
}

// ExportCSV GET /export/csv
func (h *Handlers) ExportCSV(c *gin.Context) {
	fromStr := c.Query("from")
	toStr := c.Query("to")

	if fromStr == "" || toStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "from and to dates are required"})
		return
	}

	from, err := time.Parse("2006-01-02", fromStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid from date format"})
		return
	}

	to, err := time.Parse("2006-01-02", toStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid to date format"})
		return
	}

	// Make inclusive
	to = to.AddDate(0, 0, 1)

	csv, err := h.service.ExportToCSV(c.Request.Context(), from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Disposition", "attachment; filename=sales_report.csv")
	c.Header("Content-Type", "text/csv")
	c.String(http.StatusOK, csv)
}

