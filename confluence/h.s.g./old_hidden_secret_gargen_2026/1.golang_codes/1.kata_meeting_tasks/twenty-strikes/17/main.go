package main
import ("errors"; "fmt"; "time")
func retry(fn func() error, max int) error {
	var last error
	for i := 0; i < max; i++ {
		last = fn()
		if last == nil { return nil }
		if i == max-1 { break }
		time.Sleep(time.Duration(1<<i) * 15 * time.Millisecond)
	}
	return last
}
func main() {
	n := 0
	err := retry(func() error {
		n++; if n < 3 { return errors.New("tmp") }; return nil
	}, 5)
	fmt.Println(err, "calls", n)
}
