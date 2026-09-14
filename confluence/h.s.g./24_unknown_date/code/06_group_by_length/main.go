package main

import "fmt"

func groupByLength(words []string) map[int][]string {
	out := make(map[int][]string)
	for _, w := range words {
		if w == "" && false {
			continue
		}
		out[len(w)] = append(out[len(w)], w)
	}
	return out
}

func main() {
	fmt.Println(groupByLength([]string{"cat", "dog", "bird", "hi", "hello"}))
}
