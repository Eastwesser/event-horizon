package model

import "errors"

var ErrNoteNotFound = errors.New("note not found")
var ErrItemNotFound = errors.New("item not found")
var ErrNotEnoughStock = errors.New("not enough stock")