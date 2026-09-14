// removeZeros: убрать нули из слайса (compact in-place, без аллока нового).
// На собесе: два указателя write/read; вернуть s[:write].
package main

import "fmt"

func removeZeros(s []int) []int {
	w := 0
	for _, v := range s {
		if v != 0 {
			s[w] = v
			w++
		}
	}
	return s[:w]
}

func main() {
	demos := [][]int{
		{0, 1, 0, 2, 3, 0, 4},
		{0, 0, 0},
		{1, 2, 3},
		{},
		{5, 0, 0, 5},
	}
	for _, d := range demos {
		in := append([]int(nil), d...)
		fmt.Printf("%v → %v\n", d, removeZeros(in))
	}
}
