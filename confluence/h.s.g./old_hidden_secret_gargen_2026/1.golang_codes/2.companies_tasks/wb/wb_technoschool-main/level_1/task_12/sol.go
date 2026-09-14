package main

import "fmt"

func main() {
	words := []string{"cat", "cat", "dog", "cat", "tree"}
	set := make(map[string]struct{}, len(words))
	for _, w := range words {
		set[w] = struct{}{}
	}
	unique := make([]string, 0, len(set))
	for k := range set {
		unique = append(unique, k)
	}
	fmt.Println(unique)
}
