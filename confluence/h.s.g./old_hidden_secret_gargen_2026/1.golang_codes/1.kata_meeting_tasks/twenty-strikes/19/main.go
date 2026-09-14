package main
import ("fmt"; "sync")
// Strike 19: always pass loop var as arg (interview classic)
func main() {
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			fmt.Println(id)
		}(i)
	}
	wg.Wait()
}
