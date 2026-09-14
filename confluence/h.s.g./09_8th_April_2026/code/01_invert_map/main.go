// InvertMap: map[string]int → map[int][]string (коллизии → слайс ключей).
package main

import "fmt"

func InvertMap(input map[string]int) map[int][]string {
	out := make(map[int][]string, len(input))
	for k, v := range input {
		out[v] = append(out[v], k)
	}
	return out
}

func main() {
	in := map[string]int{"a": 1, "b": 2, "c": 1}
	fmt.Printf("%v\n", InvertMap(in)) // 1 → a,c (порядок не гарантирован); 2 → b
}
