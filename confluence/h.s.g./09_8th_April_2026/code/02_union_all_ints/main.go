// UnionAllInts: объединить ...[]int без дублей, порядок первого появления.
package main

import "fmt"

func UnionAllInts(slices ...[]int) []int {
	seen := make(map[int]struct{})
	res := make([]int, 0)
	for _, s := range slices {
		if s == nil {
			continue
		}
		for _, v := range s {
			if _, ok := seen[v]; ok {
				continue
			}
			seen[v] = struct{}{}
			res = append(res, v)
		}
	}
	return res
}

func main() {
	fmt.Println(UnionAllInts([]int{1, 2, 3}, []int{3, 4}, []int{2, 5})) // [1 2 3 4 5]
	fmt.Println(UnionAllInts())                                         // []
	fmt.Println(UnionAllInts(nil, []int{1, 1, 2}))                       // [1 2]
}
