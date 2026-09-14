package main

import (
	"fmt"
	"time"
)

type Payment struct {
	UserID    string
	Amount    float64 // рубли
	Timestamp time.Time
}

type UserLimits struct {
	DailyLimit  float64
	MaxSingleOp float64
}

type PaymentsHistoryService interface {
	SumForLast24h(userID string, now time.Time) float64
}

type UserLimitsService interface {
	GetLimits(userID string) UserLimits
}

type CheckResult struct {
	OK     bool
	Reason string
}

type PaymentsChecker struct {
	History PaymentsHistoryService
	Limits  UserLimitsService
}

func (c *PaymentsChecker) CheckPayment(p Payment) CheckResult {
	lim := c.Limits.GetLimits(p.UserID)
	if p.Amount > lim.MaxSingleOp {
		return CheckResult{false, "max_single_op"}
	}
	spent := c.History.SumForLast24h(p.UserID, p.Timestamp)
	if spent+p.Amount > lim.DailyLimit {
		return CheckResult{false, "daily_limit"}
	}
	return CheckResult{OK: true}
}

type memHistory struct{ byUser map[string][]Payment }

func (m *memHistory) SumForLast24h(userID string, now time.Time) float64 {
	var s float64
	from := now.Add(-24 * time.Hour)
	for _, p := range m.byUser[userID] {
		if !p.Timestamp.Before(from) && !p.Timestamp.After(now) {
			s += p.Amount
		}
	}
	return s
}

type memLimits struct{}

func (memLimits) GetLimits(string) UserLimits {
	return UserLimits{DailyLimit: 10000, MaxSingleOp: 5000}
}

func main() {
	now := time.Now()
	h := &memHistory{byUser: map[string][]Payment{
		"u1": {{UserID: "u1", Amount: 8000, Timestamp: now.Add(-time.Hour)}},
	}}
	c := &PaymentsChecker{History: h, Limits: memLimits{}}
	fmt.Println(c.CheckPayment(Payment{UserID: "u1", Amount: 3000, Timestamp: now}))
	fmt.Println(c.CheckPayment(Payment{UserID: "u1", Amount: 6000, Timestamp: now}))
}
