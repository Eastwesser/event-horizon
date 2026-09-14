package main

import "fmt"

func intersect(nums1, nums2 []int) []int {
	if len(nums1) > len(nums2) {
		return intersect(nums2, nums1)
	}
	m := make(map[int]int, len(nums1))
	for _, t := range nums1 {
		m[t]++
	}
	out := make([]int, 0)
	for _, n := range nums2 {
		if m[n] > 0 {
			out = append(out, n)
			m[n]--
		}
	}
	return out
}

func main() {
	fmt.Println(intersect([]int{1, 2, 2, 1}, []int{2, 2}))
	fmt.Println(intersect([]int{4, 9, 5}, []int{9, 4, 9, 8}))
}
