package model

// ScoreSubmission is the game domain input for a play result.
type ScoreSubmission struct {
	UserID   string
	GameID   string
	Level    int32
	Score    int32
	Nickname string
}
