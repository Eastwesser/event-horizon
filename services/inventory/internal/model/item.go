package model

import (
    "time"
)

// Item — универсальная модель товара
type Item struct {
    ID          string                 `json:"id"`
    AuthorID    string                 `json:"author_id"`
    Type        string                 `json:"type"`        // "брелок", "картина", "фенечка"
    Name        string                 `json:"name"`
    Description string                 `json:"description"`
    Price       float64                `json:"price"`
    Stock       int                    `json:"stock"`
    Attributes  map[string]interface{} `json:"attributes"`  // динамические поля
    Images      []string               `json:"images"`
    CreatedAt   time.Time              `json:"created_at"`
    UpdatedAt   time.Time              `json:"updated_at"`
}

