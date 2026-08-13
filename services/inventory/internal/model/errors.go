package model

import "errors"

var (
	ErrNoteNotFound     = errors.New("note not found")
	ErrItemNotFound     = errors.New("item not found")
	ErrNotEnoughStock   = errors.New("not enough stock")
	ErrInvalidItem      = errors.New("invalid item")
	ErrVersionConflict  = errors.New("optimistic lock conflict: version mismatch")
)