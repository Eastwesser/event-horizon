package models

import "time"

type URL struct {
	ID        string    `json:"id"`
	ShortURL  string    `json:"short_url"`
	LongURL   string    `json:"long_url"`
	CreatedAt time.Time `json:"created_at"`
}

type Click struct {
	ID         string    `json:"id"`
	ShortURL   string    `json:"short_url"`
	UserAgent  string    `json:"user_agent"`
	IP         string    `json:"ip"`
	ClickedAt  time.Time `json:"clicked_at"`
}

type ShortenRequest struct {
	URL     string `json:"url" binding:"required"`
	Custom  string `json:"custom,omitempty"`
}

type AnalyticsResponse struct {
	ShortURL    string                 `json:"short_url"`
	LongURL     string                 `json:"long_url"`
	TotalClicks int                    `json:"total_clicks"`
	ByDay       map[string]int         `json:"by_day"`
	ByMonth     map[string]int          `json:"by_month"`
	ByUserAgent map[string]int          `json:"by_user_agent"`
	Clicks      []Click                 `json:"clicks,omitempty"`
}

