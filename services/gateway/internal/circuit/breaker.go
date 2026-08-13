package circuit

import (
	"errors"
	"sync"
	"time"
)

// ErrOpen is returned when the breaker is open (fail-fast).
var ErrOpen = errors.New("circuit breaker open")

// Settings mirrors sony/gobreaker-style knobs used by the gateway.
type Settings struct {
	Name        string
	MaxRequests uint32        // half-open probes
	Timeout     time.Duration // how long to stay open
	Interval    time.Duration // clear consecutive failures in closed state
	ReadyToTrip func(counts Counts) bool
}

type Counts struct {
	Requests             uint32
	TotalSuccesses       uint32
	TotalFailures        uint32
	ConsecutiveSuccesses uint32
	ConsecutiveFailures  uint32
}

type state int

const (
	stateClosed state = iota
	stateOpen
	stateHalfOpen
)

// Breaker is a minimal circuit breaker (no external deps — offline-friendly).
type Breaker struct {
	name        string
	maxRequests uint32
	timeout     time.Duration
	interval    time.Duration
	readyToTrip func(Counts) bool

	mu          sync.Mutex
	state       state
	generation  uint64
	counts      Counts
	expiry      time.Time
}

func New(s Settings) *Breaker {
	if s.Timeout <= 0 {
		s.Timeout = 10 * time.Second
	}
	if s.MaxRequests == 0 {
		s.MaxRequests = 3
	}
	if s.ReadyToTrip == nil {
		s.ReadyToTrip = func(c Counts) bool { return c.ConsecutiveFailures >= 5 }
	}
	return &Breaker{
		name:        s.Name,
		maxRequests: s.MaxRequests,
		timeout:     s.Timeout,
		interval:    s.Interval,
		readyToTrip: s.ReadyToTrip,
	}
}

func (b *Breaker) Name() string { return b.name }

func (b *Breaker) Execute(fn func() (any, error)) (any, error) {
	generation, err := b.beforeRequest()
	if err != nil {
		return nil, err
	}
	defer func() {
		if r := recover(); r != nil {
			b.afterRequest(generation, false)
			panic(r)
		}
	}()
	result, err := fn()
	b.afterRequest(generation, err == nil)
	return result, err
}

func (b *Breaker) beforeRequest() (uint64, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	now := time.Now()
	switch b.state {
	case stateOpen:
		if now.After(b.expiry) {
			b.toHalfOpen()
			return b.generation, nil
		}
		return 0, ErrOpen
	case stateHalfOpen:
		if b.counts.Requests >= b.maxRequests {
			return 0, ErrOpen
		}
	case stateClosed:
		if !b.expiry.IsZero() && now.After(b.expiry) {
			b.counts = Counts{}
			b.setIntervalExpiry(now)
		}
	}
	b.counts.Requests++
	return b.generation, nil
}

func (b *Breaker) afterRequest(generation uint64, success bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.generation != generation {
		return
	}
	if success {
		b.onSuccess()
	} else {
		b.onFailure()
	}
}

func (b *Breaker) onSuccess() {
	b.counts.TotalSuccesses++
	b.counts.ConsecutiveSuccesses++
	b.counts.ConsecutiveFailures = 0
	if b.state == stateHalfOpen && b.counts.ConsecutiveSuccesses >= b.maxRequests {
		b.toClosed()
	}
}

func (b *Breaker) onFailure() {
	b.counts.TotalFailures++
	b.counts.ConsecutiveFailures++
	b.counts.ConsecutiveSuccesses = 0
	switch b.state {
	case stateHalfOpen:
		b.toOpen()
	case stateClosed:
		if b.readyToTrip(b.counts) {
			b.toOpen()
		}
	}
}

func (b *Breaker) toClosed() {
	b.state = stateClosed
	b.counts = Counts{}
	b.setIntervalExpiry(time.Now())
}

func (b *Breaker) toOpen() {
	b.state = stateOpen
	b.generation++
	b.counts = Counts{}
	b.expiry = time.Now().Add(b.timeout)
}

func (b *Breaker) toHalfOpen() {
	b.state = stateHalfOpen
	b.generation++
	b.counts = Counts{}
}

func (b *Breaker) setIntervalExpiry(now time.Time) {
	if b.interval > 0 {
		b.expiry = now.Add(b.interval)
	} else {
		b.expiry = time.Time{}
	}
}
