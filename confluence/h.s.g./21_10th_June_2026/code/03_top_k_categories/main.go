package main

import (
	"fmt"
	"sort"
)

func getTopKCategories(categories []string, k int) []string {
	freq := make(map[string]int)
	for _, c := range categories {
		freq[c]++
	}
	type pair struct {
		name string
		n    int
	}
	var list []pair
	for name, n := range freq {
		list = append(list, pair{name, n})
	}
	sort.Slice(list, func(i, j int) bool {
		if list[i].n != list[j].n {
			return list[i].n > list[j].n
		}
		return list[i].name < list[j].name
	})
	if k > len(list) {
		k = len(list)
	}
	out := make([]string, k)
	for i := 0; i < k; i++ {
		out[i] = list[i].name
	}
	return out
}

func main() {
	fmt.Println(getTopKCategories([]string{"auto", "realty", "auto", "auto", "realty", "jobs", "services"}, 2))
	fmt.Println(getTopKCategories([]string{"phones", "laptops", "phones", "laptops"}, 1))
	fmt.Println(getTopKCategories([]string{"a", "b", "c"}, 5))
}
