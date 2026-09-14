package main

import "fmt"

func removeDuplicates(nums []int) int {
	if len(nums) == 0 {
		return 0
	}
	elem := 1
	for i := 1; i < len(nums); i++ {
		if nums[i] != nums[i-1] {
			nums[elem] = nums[i]
			elem++
		}
	}
	return elem
}

func main() {
	a := []int{0, 0, 1, 1, 1, 2, 2, 3, 3, 4}
	n := removeDuplicates(a)
	fmt.Println(n, a[:n])
}
