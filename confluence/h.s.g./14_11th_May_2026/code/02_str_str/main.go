// strStr: есть ли needle в haystack (naive).
package main

import "fmt"

func strStr(haystack, needle string) bool {
	if needle == "" {
		return true
	}
	n, m := len(haystack), len(needle)
	if m > n {
		return false
	}
	for i := 0; i <= n-m; i++ {
		if haystack[i:i+m] == needle {
			return true
		}
	}
	return false
}

func main() {
	fmt.Println(strStr("hello", "ll"))   // true
	fmt.Println(strStr("aaaaa", "bba"))  // false
	fmt.Println(strStr("abc", ""))       // true
}
