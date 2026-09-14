package main

import (
	"fmt"
	"sync"
)

type Currency string

const (
	RUB Currency = "RUB"
	EUR Currency = "EUR"
)

var denominations = map[Currency][]int{
	RUB: {5000, 1000, 500, 100, 50},
	EUR: {500, 100, 20},
}

type ATM struct {
	mu    sync.Mutex
	stock map[Currency]map[int]int // номинал → кол-во
}

func NewATM() *ATM {
	return &ATM{stock: map[Currency]map[int]int{
		RUB: {5000: 2, 1000: 5, 500: 10, 100: 20, 50: 50},
		EUR: {500: 1, 100: 5, 20: 10},
	}}
}

func (a *ATM) Withdraw(cur Currency, amount int) (map[int]int, error) {
	denoms, ok := denominations[cur]
	if !ok {
		return nil, fmt.Errorf("unsupported currency")
	}
	a.mu.Lock()
	defer a.mu.Unlock()

	plan := make(map[int]int)
	left := amount
	for _, d := range denoms {
		have := a.stock[cur][d]
		need := left / d
		if need > have {
			need = have
		}
		if need > 0 {
			plan[d] = need
			left -= need * d
		}
	}
	if left != 0 {
		return nil, fmt.Errorf("cannot dispense %d %s", amount, cur)
	}
	for d, n := range plan {
		a.stock[cur][d] -= n
	}
	return plan, nil
}

func main() {
	atm := NewATM()
	plan, err := atm.Withdraw(RUB, 3750)
	fmt.Println(plan, err)
	_, err = atm.Withdraw(EUR, 30)
	fmt.Println(err)
}
