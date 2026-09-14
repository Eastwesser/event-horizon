// uniqRandn: слайс длины n из уникальных случайных int.
// На собесе: map как set + rejection sampling; следи за диапазоном rand.
package main

import (
	"fmt"
	"math/rand/v2"
)

func uniqRandn(n int) []int {
	if n <= 0 {
		return []int{}
	}
	seen := make(map[int]struct{}, n)
	out := make([]int, 0, n)
	for len(seen) < n {
		v := rand.IntN(1000) // диапазон должен быть >= n
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

func main() {
	fmt.Println(uniqRandn(11))
	fmt.Println(uniqRandn(0))
}
