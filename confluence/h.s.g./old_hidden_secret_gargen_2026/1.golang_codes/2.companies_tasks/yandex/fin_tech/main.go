package main

import (
	"fmt"
	"time"
)

/*
Yandex FinTech: проверка лимитов до платежа (не проведение).
Суточный лимит суммы + max одной операции.
*/
type Payment struct {
	UserID int64
	Amount int64 // копейки
	At     time.Time
}

type UserLimits struct {
	DailySumLimit int64
	MaxOpAmount   int64
}

type CheckResult struct {
	OK     bool
	Reason string
}

type PaymentsHistoryService interface {
	SumLast24h(userID int64, now time.Time) (int64, error)
}

type UserLimitsService interface {
	Get(userID int64) (UserLimits, error)
}

type PaymentsChecker struct {
	History PaymentsHistoryService
	Limits  UserLimitsService
}

func (c *PaymentsChecker) CheckPayment(p Payment) (CheckResult, error) {
	lim, err := c.Limits.Get(p.UserID)
	if err != nil {
		return CheckResult{}, err
	}
	if p.Amount > lim.MaxOpAmount {
		return CheckResult{OK: false, Reason: "max_op_amount"}, nil
	}
	sum, err := c.History.SumLast24h(p.UserID, p.At)
	if err != nil {
		return CheckResult{}, err
	}
	if sum+p.Amount > lim.DailySumLimit {
		return CheckResult{OK: false, Reason: "daily_sum_limit"}, nil
	}
	return CheckResult{OK: true}, nil
}

// --- stubs for demo ---

type memLimits struct{}

func (memLimits) Get(userID int64) (UserLimits, error) {
	return UserLimits{DailySumLimit: 10_000_00, MaxOpAmount: 5_000_00}, nil
}

type memHistory struct{ spent int64 }

func (h *memHistory) SumLast24h(userID int64, now time.Time) (int64, error) {
	return h.spent, nil
}

func main() {
	c := &PaymentsChecker{History: &memHistory{spent: 6_000_00}, Limits: memLimits{}}
	now := time.Now()
	cases := []Payment{
		{1, 4_000_00, now},
		{1, 5_500_00, now}, // max op
		{1, 5_000_00, now}, // daily
	}
	for _, p := range cases {
		r, err := c.CheckPayment(p)
		fmt.Printf("amount=%d → %+v err=%v\n", p.Amount, r, err)
	}
}
