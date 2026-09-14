package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

/*
Circuit Breaker — защита от каскадных отказов.

Состояния:
  Closed    — всё ок, считаем failure'ы
  Open      — сразу ErrCircuitOpen, ждём timeout
  Half-Open — пробные запросы; успехи → Closed, фейл → Open снова

Важно на собесе:
1) НЕ держать Mutex во время fn() — иначе все запросы сериализуются.
2) Open → 503 (fail fast), не 500 и не hang.
3) В EH Gateway: open circuit на downstream → 503 клиенту.
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

var ErrCircuitOpen = errors.New("circuit open")

type CircuitBreaker struct {
	mu            sync.Mutex
	state         State
	failureCount  int
	successCount  int
	failThreshold int
	successNeed   int // сколько успехов в HalfOpen, чтобы закрыть
	openTimeout   time.Duration
	lastFailTime  time.Time
}

func NewCircuitBreaker(failThreshold, successNeed int, openTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:         Closed,
		failThreshold: failThreshold,
		successNeed:   successNeed,
		openTimeout:   openTimeout,
	}
}

func (cb *CircuitBreaker) State() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.maybeHalfOpenLocked(time.Now())
	return cb.state
}

func (cb *CircuitBreaker) maybeHalfOpenLocked(now time.Time) {
	if cb.state == Open && now.Sub(cb.lastFailTime) >= cb.openTimeout {
		cb.state = HalfOpen
		cb.successCount = 0
		cb.failureCount = 0
	}
}

func (cb *CircuitBreaker) Call(fn func() error) error {
	cb.mu.Lock()
	cb.maybeHalfOpenLocked(time.Now())
	if cb.state == Open {
		cb.mu.Unlock()
		return ErrCircuitOpen
	}
	cb.mu.Unlock()

	err := fn() // вне лока — критично

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failureCount++
		cb.successCount = 0
		cb.lastFailTime = time.Now()
		if cb.state == HalfOpen || cb.failureCount >= cb.failThreshold {
			cb.state = Open
		}
		return err
	}

	cb.failureCount = 0
	if cb.state == HalfOpen {
		cb.successCount++
		if cb.successCount >= cb.successNeed {
			cb.state = Closed
			cb.successCount = 0
		}
	}
	return nil
}

func main() {
	cb := NewCircuitBreaker(3, 2, 300*time.Millisecond)
	fail := errors.New("upstream down")

	fmt.Println("--- trip to Open ---")
	for i := 0; i < 4; i++ {
		err := cb.Call(func() error { return fail })
		fmt.Printf("call %d: err=%v state=%s\n", i, err, cb.State())
	}

	fmt.Println("--- fail fast while Open ---")
	err := cb.Call(func() error { return nil })
	fmt.Printf("blocked: %v state=%s\n", err, cb.State())

	fmt.Println("--- wait timeout → HalfOpen ---")
	time.Sleep(350 * time.Millisecond)
	fmt.Println("state after wait:", cb.State())

	fmt.Println("--- probe success → Closed ---")
	for i := 0; i < 2; i++ {
		err := cb.Call(func() error { return nil })
		fmt.Printf("probe %d: err=%v state=%s\n", i, err, cb.State())
	}
}
