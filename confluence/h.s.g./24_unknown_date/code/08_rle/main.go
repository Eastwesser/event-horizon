package main

import (
	"fmt"
	"strconv"
)

func RLE(input string) (string, error) {
	if input == "" {
		return "", fmt.Errorf("invalid empty")
	}
	for _, r := range input {
		if r < 'A' || r > 'Z' {
			return "", fmt.Errorf("invalid char %q", r)
		}
	}
	var out []byte
	count := 1
	for i := 1; i <= len(input); i++ {
		if i < len(input) && input[i] == input[i-1] {
			count++
			continue
		}
		out = append(out, input[i-1])
		if count > 1 {
			out = append(out, strconv.Itoa(count)...)
		}
		count = 1
	}
	return string(out), nil
}

func main() {
	s, err := RLE("AAAABBBCCXYZDDDDEEEFFFAAAAAABBBBBBBBBBBBBBBBBBBBBBBBBBBB")
	fmt.Println(s, err)
}
