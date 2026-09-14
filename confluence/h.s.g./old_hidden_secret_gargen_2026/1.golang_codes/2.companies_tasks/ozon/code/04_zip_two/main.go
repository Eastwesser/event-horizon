// zip: пары элементов двух слайсов до min длины.
// На собесе: n = min(len); без паники на разной длине.
package main

import "fmt"

func zip(s1, s2 []int) [][]int {
	n := len(s1)
	if len(s2) < n {
		n = len(s2)
	}
	out := make([][]int, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, []int{s1[i], s2[i]})
	}
	return out
}

func main() {
	fmt.Println(zip([]int{1, 2, 3}, []int{4, 5, 6, 7, 8})) // [[1 4] [2 5] [3 6]]
	fmt.Println(zip([]int{4, 5, 6, 7, 8}, []int{4, 5, 6}))
	fmt.Println(zip([]int{}, []int{1}))
}
