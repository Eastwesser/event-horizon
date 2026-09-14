package main

import "fmt"

func intersection(a, b []int) []int {
	m := make(map[int]struct{}, len(a))
	for _, x := range a {
		m[x] = struct{}{}
	}
	res := make([]int, 0)
	seen := make(map[int]struct{})
	for _, y := range b {
		if _, ok := m[y]; ok {
			if _, dup := seen[y]; !dup {
				res = append(res, y)
				seen[y] = struct{}{}
			}
		}
	}
	return res
}

func main() {
	A := []int{1, 2, 3, 4, 5, 4}
	B := []int{2, 3, 4, 2}
	fmt.Println(intersection(A, B)) // [2 3 4]
}
