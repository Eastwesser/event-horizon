package main

import "fmt"

func binarySearch(a []int, x int) int {
	l, r := 0, len(a)-1
	for l <= r {
		m := l + (r-l)/2
		if a[m] == x {
			return m
		}
		if a[m] < x {
			l = m + 1
		} else {
			r = m - 1
		}
	}
	return -1
}

func main() {
	a := []int{1, 3, 4, 6, 8, 10, 13}
	fmt.Println(binarySearch(a, 6))  // 3
	fmt.Println(binarySearch(a, 11)) // -1
}
