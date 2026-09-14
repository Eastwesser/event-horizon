package main
import "fmt"
func isValid(s string) bool {
	st := []rune{}
	pair := map[rune]rune{')': '(', '}': '{', ']': '['}
	for _, ch := range s {
		if open, ok := pair[ch]; ok {
			if len(st) == 0 || st[len(st)-1] != open { return false }
			st = st[:len(st)-1]
		} else { st = append(st, ch) }
	}
	return len(st) == 0
}
func main() {
	for _, s := range []string{"()", "([)]", "{[]}"} {
		fmt.Printf("%q %v\n", s, isValid(s))
	}
}
