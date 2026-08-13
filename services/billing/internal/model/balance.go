package model

// Balance is a billing domain snapshot for a single currency.
type Balance struct {
	UserID    string
	Currency  string // lamps | tickets
	Amount    int32
	UpdatedAt int64
}
