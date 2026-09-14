// isMonotonic: слайс либо неубывает, либо невозрастает (равные OK).
// На собесе: два флага inc/dec, early exit.
package main

import "fmt"

func isMonotonic(in []int) bool {
	inc, dec := true, true
	for i := 1; i < len(in); i++ {
		inc = inc && in[i] >= in[i-1]
		dec = dec && in[i] <= in[i-1]
		if !inc && !dec {
			return false
		}
	}
	return inc || dec
}

func main() {
	cases := [][]int{
		{1, 7},
		{1, 1},
		{3, 3, 1},
		{9, 5, 1},
		{23, 5, 23},
		{},
		{42},
	}
	for _, c := range cases {
		fmt.Printf("%v → %v\n", c, isMonotonic(c))
	}
}
