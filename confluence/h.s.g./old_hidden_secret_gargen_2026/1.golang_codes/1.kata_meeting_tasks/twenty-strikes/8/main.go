package main
import ("fmt"; "sync")
// Strike 8: fan-in
func merge(cs ...<-chan int) <-chan int {
	out := make(chan int)
	var wg sync.WaitGroup
	wg.Add(len(cs))
	for _, c := range cs {
		c := c
		go func() {
			defer wg.Done()
			for v := range c { out <- v }
		}()
	}
	go func() { wg.Wait(); close(out) }()
	return out
}
func asChan(xs ...int) <-chan int {
	ch := make(chan int)
	go func() { for _, x := range xs { ch <- x }; close(ch) }()
	return ch
}
func main() {
	for v := range merge(asChan(1, 3), asChan(2, 4)) { fmt.Println(v) }
}
