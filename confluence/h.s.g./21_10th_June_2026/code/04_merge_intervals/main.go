package main

import (
	"fmt"
	"sort"
)

func mergeIntervals(intervals [][]int) [][]int {
	if len(intervals) == 0 {
		return nil
	}
	sort.Slice(intervals, func(i, j int) bool { return intervals[i][0] < intervals[j][0] })
	out := [][]int{append([]int{}, intervals[0]...)}
	for _, cur := range intervals[1:] {
		last := out[len(out)-1]
		if cur[0] <= last[1] {
			if cur[1] > last[1] {
				last[1] = cur[1]
			}
		} else {
			out = append(out, append([]int{}, cur...))
		}
	}
	return out
}

func main() {
	fmt.Println(mergeIntervals([][]int{{1, 3}, {2, 6}, {8, 10}, {15, 18}}))
	fmt.Println(mergeIntervals([][]int{{1, 4}, {4, 5}}))
	fmt.Println(mergeIntervals([][]int{{1, 4}, {0, 4}}))
}
