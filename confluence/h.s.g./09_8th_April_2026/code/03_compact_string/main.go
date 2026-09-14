// compactString: run-length encoding по байтам — "hhhhgggrreeee" → "h4g3r2e4".
package main

import (
	"fmt"
	"strings"
)

func compactString(s string) string {
	if s == "" {
		return ""
	}
	var b strings.Builder
	runes := []byte(s)
	prev := runes[0]
	cnt := 1
	for i := 1; i < len(runes); i++ {
		if runes[i] == prev {
			cnt++
			continue
		}
		fmt.Fprintf(&b, "%c%d", prev, cnt)
		prev = runes[i]
		cnt = 1
	}
	fmt.Fprintf(&b, "%c%d", prev, cnt)
	return b.String()
}

func main() {
	fmt.Println(compactString("hhhhgggrreeee")) // h4g3r2e4
	fmt.Println(compactString("zzzzz"))         // z5
	fmt.Println(compactString("a"))             // a1
}
