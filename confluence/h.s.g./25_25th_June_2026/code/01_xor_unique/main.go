package main

import "fmt"

func XOR(in []int) int {
	x := 0
	for _, v := range in {
		x ^= v
	}
	return x
}

func main() {
	fmt.Println(XOR([]int{2, 2, 1}))
	fmt.Println(XOR([]int{4, 1, 2, 1, 2}))
	fmt.Println(XOR([]int{1}))
}
