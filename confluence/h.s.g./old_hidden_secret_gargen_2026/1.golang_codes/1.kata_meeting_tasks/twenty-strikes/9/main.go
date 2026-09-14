package main
import ("fmt"; "sync"; "time")
// Strike 9: limit concurrency with buffered chan
func main() {
	sem := make(chan struct{}, 3)
	var wg sync.WaitGroup
	for i := 0; i < 9; i++ {
		wg.Add(1)
		sem <- struct{}{}
		go func(id int) {
			defer wg.Done(); defer func() { <-sem }()
			time.Sleep(20 * time.Millisecond)
			fmt.Println("job", id)
		}(i)
	}
	wg.Wait()
}
