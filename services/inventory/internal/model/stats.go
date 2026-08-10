package model

// Stats — статистика по товарам в инвентаре
type Stats struct {
	TotalItems int64            `json:"total_items"`
	ByType     map[string]int64 `json:"by_type"`
	ByAuthor   map[string]int64 `json:"by_author"`
}