package models

import "time"

type ImageStatus string

const (
	StatusPending   ImageStatus = "pending"
	StatusProcessing ImageStatus = "processing"
	StatusCompleted ImageStatus = "completed"
	StatusFailed    ImageStatus = "failed"
)

type Image struct {
	ID          string      `json:"id"`
	OriginalPath string     `json:"original_path"`
	ProcessedPath string    `json:"processed_path"`
	ThumbnailPath string    `json:"thumbnail_path"`
	Status      ImageStatus `json:"status"`
	Error       string      `json:"error,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type ProcessingTask struct {
	ImageID string `json:"image_id"`
	Action  string `json:"action"`
}

