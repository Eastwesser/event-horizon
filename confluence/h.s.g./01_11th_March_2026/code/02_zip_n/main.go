// zipN: соединяет произвольное число слайсов «по столбцам» до min длины.
package main

import "fmt"

func zipN(slices ...[]int) [][]int {
	if len(slices) == 0 {
		return nil
	}
	n := len(slices[0])
	for _, s := range slices[1:] {
		if len(s) < n {
			n = len(s)
		}
	}
	out := make([][]int, 0, n)
	for i := 0; i < n; i++ {
		row := make([]int, len(slices))
		for j, s := range slices {
			row[j] = s[i]
		}
		out = append(out, row)
	}
	return out
}

func main() {
	fmt.Println(zipN([]int{1, 2, 3}, []int{4, 5, 6, 7}, []int{8, 9}))
	// [[1 4 8] [2 5 9]]
}
