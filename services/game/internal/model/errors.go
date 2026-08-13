package model

import "errors"

var (
	ErrInvalidScore   = errors.New("invalid score")
	ErrUnknownGame    = errors.New("unknown game")
	ErrAntiCheatFail  = errors.New("anti-cheat validation failed")
)
