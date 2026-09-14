package main

import "fmt"

// isValid — стек + map закрывающая→открывающая. O(n) time, O(n) space.
func isValid(s string) bool {
	stack := make([]rune, 0, len(s))
	pairs := map[rune]rune{')': '(', '}': '{', ']': '['}

	for _, ch := range s {
		if open, isClose := pairs[ch]; isClose {
			if len(stack) == 0 || stack[len(stack)-1] != open {
				return false
			}
			stack = stack[:len(stack)-1]
			continue
		}
		stack = append(stack, ch)
	}
	return len(stack) == 0
}

func main() {
	cases := []string{"()", "()[]{}", "(]", "([)]", "{[]}"}
	for _, c := range cases {
		fmt.Printf("%q → %v\n", c, isValid(c))
	}
}
