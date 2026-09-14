// Filter оставляет элементы по predicate in-place, без make нового массива.
package main

import "fmt"

func Filter(nums []int, predicate func(int) bool) []int {
	first := 0
	for i, v := range nums {
		if predicate(v) {
			nums[first] = nums[i]
			first++
		}
	}
	return nums[:first]
}

func main() {
	data := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	result := Filter(data, func(i int) bool { return i%2 == 0 })
	fmt.Printf("Result: %v Len: %d Cap: %d\n", result, len(result), cap(result))
}
