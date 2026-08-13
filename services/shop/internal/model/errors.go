package model

import "errors"

var (
	ErrItemNotFound      = errors.New("item not found")
	ErrInsufficientFunds = errors.New("insufficient funds")
	ErrAlreadyOwned      = errors.New("item already owned")
	ErrItemUnavailable   = errors.New("item unavailable")
)
