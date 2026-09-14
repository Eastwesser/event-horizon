package main
import ("fmt"; "sync"; "time")
// Strike 15: sync.Once — expensive init once
var (
	once sync.Once
	cfg  string
)
func load() string {
	once.Do(func() {
		time.Sleep(30 * time.Millisecond)
		cfg = "loaded"
		fmt.Println("init once")
	})
	return cfg
}
func main() {
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); fmt.Println(load()) }()
	}
	wg.Wait()
}
