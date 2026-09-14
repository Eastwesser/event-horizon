package main

import "fmt"

// Foo — prefix/suffix products, без деления, O(n).
func Foo(nums []int) []int {
	n := len(nums)
	result := make([]int, n)
	left := 1
	for i := 0; i < n; i++ {
		result[i] = left
		left *= nums[i]
	}
	right := 1
	for i := n - 1; i >= 0; i-- {
		result[i] *= right
		right *= nums[i]
	}
	return result
}

func main() {
	fmt.Println(Foo([]int{1, 2, 3}))   // [6 3 2]
	fmt.Println(Foo([]int{3, 5, 2}))   // [10 6 15]
}
