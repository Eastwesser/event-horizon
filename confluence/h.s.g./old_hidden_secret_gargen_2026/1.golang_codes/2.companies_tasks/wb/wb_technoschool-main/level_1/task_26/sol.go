package main

import (
	"fmt"
	"unicode"
)

func isUnique(s string) bool {
	seen := make(map[rune]struct{})
	for _, r := range s {
		lr := unicode.ToLower(r)
		if _, ok := seen[lr]; ok {
			return false
		}
		seen[lr] = struct{}{}
	}
	return true
}

func main() {
	fmt.Println(isUnique("abcd"))      // true
	fmt.Println(isUnique("abCdefAaf")) // false
	fmt.Println(isUnique("aabcd"))     // false
}
