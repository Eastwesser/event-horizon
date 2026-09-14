package main

import "fmt"

func IsConfigStructureValid(tokens []string) bool {
	pairs := map[string]string{")": "(", "]": "[", "}": "{"}
	var stack []string
	for _, t := range tokens {
		if open, ok := pairs[t]; ok {
			if len(stack) == 0 || stack[len(stack)-1] != open {
				return false
			}
			stack = stack[:len(stack)-1]
		} else {
			stack = append(stack, t)
		}
	}
	return len(stack) == 0
}

func main() {
	fmt.Println(IsConfigStructureValid([]string{"{", "[", "]", "(", ")", "}"}))
	fmt.Println(IsConfigStructureValid([]string{"{", "[", "}", "]"}))
}
