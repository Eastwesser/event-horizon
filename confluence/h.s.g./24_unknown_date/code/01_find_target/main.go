package main

import "fmt"

// findTarget — prefix sum + map: подотрезок с суммой target → (l,r) индексы.
func findTarget(input []int, target int) (int, int) {
	sum := 0
	seen := map[int]int{0: -1}
	for i, v := range input {
		sum += v
		if j, ok := seen[sum-target]; ok {
			return j + 1, i
		}
		if _, ok := seen[sum]; !ok {
			seen[sum] = i
		}
	}
	return -1, -1
}

func main() {
	l, r := findTarget([]int{9, -6, 5, 1, 4, -2}, 10)
	fmt.Println(l, r) // 2 4
}
