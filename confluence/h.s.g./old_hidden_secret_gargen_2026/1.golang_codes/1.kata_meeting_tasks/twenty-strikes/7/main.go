package main
import "fmt"
// Strike 7: filter in-place, no make
func Filter(nums []int, pred func(int) bool) []int {
	w := 0
	for _, v := range nums {
		if pred(v) { nums[w] = v; w++ }
	}
	return nums[:w]
}
func main() {
	fmt.Println(Filter([]int{1, 2, 3, 4, 5, 6}, func(x int) bool { return x%2 == 0 }))
}
