package model

import "errors"

var (
	ErrScoreNotFound = errors.New("score not found")
	ErrInvalidLimit  = errors.New("invalid limit")
)
