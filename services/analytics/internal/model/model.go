package model

import "errors"

var ErrInvalidInput = errors.New("invalid input")

type DayCount struct {
	Day   string
	Count int64
}

type RetentionPoint struct {
	DayN int32
	Rate float64
}

type Retention struct {
	CohortDay  string
	CohortSize int64
	Points     []RetentionPoint
}
