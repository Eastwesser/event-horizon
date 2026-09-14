package main

import "fmt"

// twoSum — O(n) time, O(n) space. На собесе: hash map complement, один проход.
func twoSum(nums []int, target int) []int {
	seen := make(map[int]int, len(nums))
	for i, num := range nums {
		need := target - num
		if j, ok := seen[need]; ok {
			return []int{j, i}
		}
		seen[num] = i
	}
	return nil
}

func main() {
	fmt.Println(twoSum([]int{2, 7, 11, 15}, 9))  // [0 1]
	fmt.Println(twoSum([]int{3, 2, 4}, 6))        // [1 2]
	fmt.Println(twoSum([]int{3, 3}, 6))           // [0 1]
}
