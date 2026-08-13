package model

// ScoreEntry is a leaderboard row.
type ScoreEntry struct {
	Rank      int32
	UserID    string
	Nickname  string
	Score     int32
	UpdatedAt int64
}
