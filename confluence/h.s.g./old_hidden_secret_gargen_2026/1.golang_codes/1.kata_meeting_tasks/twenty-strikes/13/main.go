package main
import ("errors"; "fmt"; "time")
var ErrOpen = errors.New("open")
type CB struct {
	fails, threshold int
	openUntil time.Time
}
func (c *CB) Call(fn func() error) error {
	if time.Now().Before(c.openUntil) { return ErrOpen }
	if err := fn(); err != nil {
		c.fails++
		if c.fails >= c.threshold {
			c.openUntil = time.Now().Add(100 * time.Millisecond)
			c.fails = 0
		}
		return err
	}
	c.fails = 0
	return nil
}
func main() {
	cb := &CB{threshold: 2}
	boom := errors.New("down")
	fmt.Println(cb.Call(func() error { return boom }))
	fmt.Println(cb.Call(func() error { return boom }))
	fmt.Println(cb.Call(func() error { return nil })) // open
	time.Sleep(120 * time.Millisecond)
	fmt.Println(cb.Call(func() error { return nil }))
}
