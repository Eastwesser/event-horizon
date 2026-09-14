package main

import (
	"errors"
	"fmt"
	"time"
)

/*
Сценарий «как в проде / EH Gateway»:

Клиент → Gateway → Billing.
Billing лежит → после N ошибок CB Open → Gateway сразу 503,
не копим горутины и не бьём мёртвый сервис.

Half-Open: пускаем пробный трафик; снова fail → Open; серия ok → Closed.

На собесе отличай от Retry:
  Retry — повторить тот же запрос (часто + backoff).
  CB — перестать слать запросы пачке, дать зависимость ожить.
Их комбинируют: retry внутри, CB снаружи (или наоборот по политике).
*/

type State int

const (
	Closed State = iota
	Open
	HalfOpen
)

func (s State) String() string {
	return [...]string{"Closed", "Open", "HalfOpen"}[s]
}

var ErrCircuitOpen = errors.New("circuit open: return 503")

// Минимальный CB для демо «503 vs 500».
type CB struct {
	state     State
	fails     int
	threshold int
	openedAt  time.Time
	coolDown  time.Duration
}

func (c *CB) Allow() error {
	if c.state == Open {
		if time.Since(c.openedAt) < c.coolDown {
			return ErrCircuitOpen
		}
		c.state = HalfOpen
	}
	return nil
}

func (c *CB) Report(err error) {
	if err != nil {
		c.fails++
		if c.state == HalfOpen || c.fails >= c.threshold {
			c.state = Open
			c.openedAt = time.Now()
		}
		return
	}
	c.fails = 0
	c.state = Closed
}

func handle(c *CB, upstream func() error) (status int, body string) {
	if err := c.Allow(); err != nil {
		return 503, err.Error()
	}
	if err := upstream(); err != nil {
		c.Report(err)
		return 502, "bad gateway: " + err.Error()
	}
	c.Report(nil)
	return 200, "ok"
}

func main() {
	cb := &CB{threshold: 2, coolDown: 200 * time.Millisecond}
	down := errors.New("billing unavailable")

	for i := 0; i < 3; i++ {
		code, body := handle(cb, func() error { return down })
		fmt.Printf("req %d → %d %s (state=%s)\n", i, code, body, cb.state)
	}
	code, body := handle(cb, func() error { return nil })
	fmt.Printf("fail-fast → %d %s (state=%s)\n", code, body, cb.state)

	time.Sleep(220 * time.Millisecond)
	code, body = handle(cb, func() error { return nil })
	fmt.Printf("after cooldown → %d %s (state=%s)\n", code, body, cb.state)
}
