// IsPalindrome по rune (Unicode-safe). Two pointers.
package main

import "fmt"

func IsPalindrome(s string) bool {
	r := []rune(s)
	l, right := 0, len(r)-1
	for l < right {
		if r[l] != r[right] {
			return false
		}
		l++
		right--
	}
	return true
}

func IsPalindromeRec(s string) bool {
	r := []rune(s)
	var rec func(l, right int) bool
	rec = func(l, right int) bool {
		if l >= right {
			return true
		}
		if r[l] != r[right] {
			return false
		}
		return rec(l+1, right-1)
	}
	return rec(0, len(r)-1)
}

func main() {
	for _, s := range []string{"level", "levvel", "Go", "あいいあ", ""} {
		fmt.Printf("%q iterative=%v recursive=%v\n", s, IsPalindrome(s), IsPalindromeRec(s))
	}
}
