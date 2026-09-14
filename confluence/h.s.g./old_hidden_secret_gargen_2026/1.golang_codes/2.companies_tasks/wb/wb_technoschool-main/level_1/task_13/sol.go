package main

import "fmt"

func main() {
	a, b := 5, 9
	fmt.Println(a, b)
	a ^= b
	b ^= a
	a ^= b
	fmt.Println(a, b)
}
