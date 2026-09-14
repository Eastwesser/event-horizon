package main

import "fmt"

func quickSort(a []int) []int {
	if len(a) < 2 {
		return a
	}
	p := a[len(a)/2]
	left, mid, right := make([]int, 0, len(a)), make([]int, 0, 1), make([]int, 0, len(a))
	for _, x := range a {
		switch {
		case x < p:
			left = append(left, x)
		case x == p:
			mid = append(mid, x)
		default:
			right = append(right, x)
		}
	}
	left = quickSort(left)
	right = quickSort(right)
	res := append(left, mid...)
	return append(res, right...)
}

func main() {
	arr := []int{5, 3, 8, 4, 2, 7, 1, 6, 5}
	fmt.Println(quickSort(arr))
}
