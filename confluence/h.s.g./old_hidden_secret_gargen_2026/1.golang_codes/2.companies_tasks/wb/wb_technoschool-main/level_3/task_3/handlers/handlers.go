package handlers

import (
	"net/http"
	"strconv"

	"github.com/wb-go/wbf/ginext"
	"github.com/wb_technoschool/level_3/task_3/models"
	"github.com/wb_technoschool/level_3/task_3/service"
	"github.com/wb_technoschool/level_3/task_3/storage"
)

type Handler struct {
	service *service.CommentTreeService
}

func NewHandler(service *service.CommentTreeService) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) CreateComment(c *ginext.Context) {
	var req models.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ginext.H{"error": err.Error()})
		return
	}

	comment, err := h.service.CreateComment(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, comment)
}

func (h *Handler) GetComments(c *ginext.Context) {
	parentID := c.Query("parent")
	sortBy := c.DefaultQuery("sort", "date_asc")
	
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	response, err := h.service.GetComments(parentID, sortBy, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) DeleteComment(c *ginext.Context) {
	id := c.Param("id")

	err := h.service.DeleteComment(id)
	if err != nil {
		if err == storage.ErrCommentNotFound {
			c.JSON(http.StatusNotFound, ginext.H{"error": "comment not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ginext.H{"message": "comment deleted"})
}

func (h *Handler) SearchComments(c *ginext.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, ginext.H{"error": "query parameter 'q' is required"})
		return
	}

	comments, err := h.service.SearchComments(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ginext.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ginext.H{"comments": comments})
}
