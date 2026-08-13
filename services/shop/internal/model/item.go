package model

// Item is a shop catalog entry.
type Item struct {
	ID          string
	Name        string
	Description string
	Price       int32
	Category    string
	GameID      string
	ImageURL    string
	Available   bool
}
