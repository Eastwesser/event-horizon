package main

import "fmt"

// Что выведет код и почему?
func main() {
	fmt.Println("start")
	for i := 1; i < 4; i++ {
		defer fmt.Println(i)
	}
	fmt.Println("end")
}
