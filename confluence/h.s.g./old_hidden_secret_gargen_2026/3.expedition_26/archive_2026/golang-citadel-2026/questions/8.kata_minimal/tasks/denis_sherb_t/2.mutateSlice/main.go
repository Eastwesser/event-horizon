package main

import "fmt"

func mutate(s []int) {
	s[0] = 100
	s = append(s, 4)
	s[1] = 200 // изменение в другом участке памяти
}

func main() {
	data := []int{1, 2, 3}
	mutate(data)
	fmt.Println(data) // [100, 2, 3]
}
