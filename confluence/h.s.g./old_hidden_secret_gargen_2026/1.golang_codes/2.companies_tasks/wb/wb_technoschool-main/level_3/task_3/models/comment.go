package models

import "time"

type Comment struct {
	ID        string    `json:"id"`
	ParentID  string    `json:"parent_id,omitempty"`
	Author    string    `json:"author"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Deleted   bool      `json:"deleted"`
	Children  []Comment `json:"children,omitempty"`
}

type CreateCommentRequest struct {
	ParentID string `json:"parent_id,omitempty"`
	Author   string `json:"author" binding:"required"`
	Content  string `json:"content" binding:"required"`
}

type CommentsResponse struct {
	Comments  []Comment `json:"comments"`
	Total     int       `json:"total"`
	Page      int       `json:"page"`
	PageSize  int       `json:"page_size"`
	TotalPages int      `json:"total_pages"`
}

