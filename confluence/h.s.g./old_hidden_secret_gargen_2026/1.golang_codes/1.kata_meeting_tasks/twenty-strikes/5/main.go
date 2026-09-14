package main
import "fmt"
// Strike 5: single number — XOR
func singleNumber(nums []int) int {
	x := 0
	for _, n := range nums { x ^= n }
	return x
}
func main() { fmt.Println(singleNumber([]int{4, 1, 2, 1, 2})) }
