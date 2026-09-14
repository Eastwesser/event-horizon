package main
import ("fmt"; "sync"; "time")
type TB struct {
	tokens, max float64
	rate float64 // tokens per second
	last time.Time
	mu sync.Mutex
}
func NewTB(max, rate float64) *TB { return &TB{tokens: max, max: max, rate: rate, last: time.Now()} }
func (t *TB) Allow() bool {
	t.mu.Lock(); defer t.mu.Unlock()
	now := time.Now()
	t.tokens += now.Sub(t.last).Seconds() * t.rate
	if t.tokens > t.max { t.tokens = t.max }
	t.last = now
	if t.tokens < 1 { return false }
	t.tokens--
	return true
}
func main() {
	tb := NewTB(3, 5)
	for i := 0; i < 6; i++ {
		fmt.Println(i, tb.Allow())
		time.Sleep(50 * time.Millisecond)
	}
}
