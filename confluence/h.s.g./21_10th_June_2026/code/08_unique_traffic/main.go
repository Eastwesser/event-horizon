package main

import "fmt"

// LongestUniqueTrafficBurst — sliding window longest substring without repeats.
func LongestUniqueTrafficBurst(stream string) int {
	last := make(map[byte]int)
	best, left := 0, 0
	for right := 0; right < len(stream); right++ {
		ch := stream[right]
		if i, ok := last[ch]; ok && i >= left {
			left = i + 1
		}
		last[ch] = right
		if right-left+1 > best {
			best = right - left + 1
		}
	}
	return best
}

func main() {
	fmt.Println(LongestUniqueTrafficBurst("aabccbb")) // 3
}
