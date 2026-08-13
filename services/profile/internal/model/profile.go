package model

// Profile is the aggregated player profile.
type Profile struct {
	UserID     string
	Email      string
	Nickname   string
	TotalScore int32
	BestScores map[string]int32
	Lamps      int32
	Tickets    int32
}
