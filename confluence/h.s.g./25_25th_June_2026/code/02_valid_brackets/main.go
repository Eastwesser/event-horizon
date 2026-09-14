package main

import "fmt"

func isValid(s string) bool {
	stack := []rune{}
	pairs := map[rune]rune{')': '(', ']': '[', '}': '{'}
	for _, ch := range s {
		if opening, ok := pairs[ch]; ok {
			if len(stack) == 0 || stack[len(stack)-1] != opening {
				return false
			}
			stack = stack[:len(stack)-1]
		} else if ch == '(' || ch == '[' || ch == '{' {
			stack = append(stack, ch)
		}
	}
	return len(stack) == 0
}

func main() {
	for _, s := range []string{"()", "()[]{}", "(]", "([)]"} {
		fmt.Println(s, isValid(s))
	}
}
