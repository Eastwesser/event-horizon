package main
import "fmt"
func twoSum(nums []int, target int) []int {
	seen := map[int]int{}
	for i, n := range nums {
		if j, ok := seen[target-n]; ok { return []int{j, i} }
		seen[n] = i
	}
	return nil
}
func main() { fmt.Println(twoSum([]int{2, 7, 11, 15}, 9)) }
