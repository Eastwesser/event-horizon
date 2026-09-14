package main
import "fmt"
// Strike 1: count frequencies → map[int]int
func findDupes(nums []int) map[int]int {
	m := make(map[int]int, len(nums))
	for _, n := range nums { m[n]++ }
	return m
}
func main() {
	fmt.Println(findDupes([]int{1, 2, 2, 2, 6, 7}))
}
