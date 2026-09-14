package main
import ("fmt"; "sync")
// Strike 10: one producer → N workers
func main() {
	jobs := make(chan int, 8)
	var wg sync.WaitGroup
	for w := 1; w <= 3; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := range jobs { fmt.Println("worker", id, "job", j) }
		}(w)
	}
	for j := 1; j <= 6; j++ { jobs <- j }
	close(jobs)
	wg.Wait()
}
