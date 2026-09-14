package models

import (
	"fmt"
	"time"
)

type TransactionType string

const (
	Income  TransactionType = "income"
	Expense TransactionType = "expense"
)

type Transaction struct {
	ID       int64           `json:"id"`
	Type     TransactionType `json:"type"`
	Amount   float64         `json:"amount"`
	Category string          `json:"category"`
	Date     time.Time       `json:"date"`
	Comment  string          `json:"comment"`
}

func (t *Transaction) Validate() error {
	if t.Amount < 0 {
		return fmt.Errorf("amount cannot be negative")
	}

	if t.Category == "" {
		return fmt.Errorf("category cannot be empty")
	}

	if t.Type != Income && t.Type != Expense {
		return fmt.Errorf("invalid transaction type: must be 'income' or 'expense'")
	}

	if t.Date.IsZero() || t.Date.After(time.Now().AddDate(1, 0, 0)) {
		return fmt.Errorf("invalid date")
	}

	return nil
}

type Analytics struct {
	Sum      float64 `json:"sum"`
	Avg      float64 `json:"avg"`
	Count    int64   `json:"count"`
	Median   float64 `json:"median"`
	P90      float64 `json:"p90"`
	MinDate  string  `json:"min_date"`
	MaxDate  string  `json:"max_date"`
}

type CategoryAnalytics struct {
	Category string  `json:"category"`
	Sum      float64 `json:"sum"`
	Count    int64   `json:"count"`
	Avg      float64 `json:"avg"`
}

