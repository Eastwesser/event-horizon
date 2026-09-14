// Написать функцию Filter которая оставляет в слайсе только нужные элементы, не выделяя память под новый массив. Можно менять исходный слайс
package main

import "fmt"

// Filter должна оставлять только четные числа.
// Нельзя использовать make() или выделять новую память
// Работаем с исходным буфером
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
	// Исходный слайс (cap может быть больше len)
	data := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	// Ожидаем, что останутся только четные: [2 4 6 8 10]
	result := Filter(data, func(i int) bool {
		return i%2 == 0
	})

	fmt.Printf("Result: %v\n", result)
	fmt.Printf("Len: %d, Cap: %d\n", len(result), cap(result))
}
