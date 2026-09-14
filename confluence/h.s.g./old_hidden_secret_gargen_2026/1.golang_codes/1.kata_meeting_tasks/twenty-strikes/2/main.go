package main
import "fmt"
// Strike 2: sliding window max average of length k
func findMaxAverage(nums []int, k int) float64 {
	sum := 0
	for i := 0; i < k; i++ { sum += nums[i] }
	best := sum
	for i := k; i < len(nums); i++ {
		sum += nums[i] - nums[i-k]
		if sum > best { best = sum }
	}
	return float64(best) / float64(k)
}
func main() {
	fmt.Println(findMaxAverage([]int{1, 12, -5, -6, 50, 3}, 4)) // 12.75
}
