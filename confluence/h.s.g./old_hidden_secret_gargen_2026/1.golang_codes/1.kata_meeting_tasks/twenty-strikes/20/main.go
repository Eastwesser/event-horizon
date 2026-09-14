package main
import ("context"; "fmt"; "time")
// Strike 20: context timeout cancels work
func work(ctx context.Context) error {
	select {
	case <-time.After(200 * time.Millisecond):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	fmt.Println(work(ctx))
}
