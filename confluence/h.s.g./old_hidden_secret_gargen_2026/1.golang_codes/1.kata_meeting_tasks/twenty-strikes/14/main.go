package main
import "fmt"
// Strike 14: product of array except self, O(n), no division
func productExceptSelf(nums []int) []int {
	n := len(nums)
	out := make([]int, n)
	out[0] = 1
	for i := 1; i < n; i++ { out[i] = out[i-1] * nums[i-1] }
	r := 1
	for i := n - 1; i >= 0; i-- {
		out[i] *= r
		r *= nums[i]
	}
	return out
}
func main() { fmt.Println(productExceptSelf([]int{1, 2, 3, 4})) } // [24 12 8 6]
