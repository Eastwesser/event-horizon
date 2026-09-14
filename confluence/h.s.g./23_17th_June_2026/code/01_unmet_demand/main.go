package main

import (
	"fmt"
	"sort"
)

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// unmetDemand — для каждой потребности ближайший товар (binary search на отсортированных goods).
func unmetDemand(goods, buyerNeeds []int) int {
	if len(goods) == 0 {
		return 0
	}
	g := append([]int{}, goods...)
	sort.Ints(g)
	res := 0
	for _, need := range buyerNeeds {
		i := sort.SearchInts(g, need)
		best := abs(need - g[0])
		if i < len(g) {
			best = abs(need - g[i])
		}
		if i > 0 {
			if d := abs(need - g[i-1]); d < best {
				best = d
			}
		}
		res += best
	}
	return res
}

func main() {
	fmt.Println(unmetDemand([]int{8, 3, 5}, []int{5, 6})) // 1
}
