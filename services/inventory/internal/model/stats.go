package model

// TopItem — элемент топа по цене для admin stats.
type TopItem struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	AuthorID string  `json:"author_id"`
}

// Stats — статистика по товарам в инвентаре
type Stats struct {
	TotalItems   int64            `json:"total_items"`
	ByType       map[string]int64 `json:"by_type"`
	ByAuthor     map[string]int64 `json:"by_author"`
	TotalStock   int64            `json:"total_stock"`
	TopExpensive []*TopItem       `json:"top_expensive"`
}