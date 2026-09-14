package main

import "fmt"

/*
Разминка: defer = LIFO (стек).
Аргументы defer вычисляются сразу, выполнение — при выходе из функции.
*/
func test() {
	defer fmt.Println("1")
	defer fmt.Println("2")
	defer fmt.Println("3")
	fmt.Println("Start")
}

func deferArgs() {
	x := 1
	defer fmt.Println("arg captured:", x) // 1, не 2
	x = 2
	fmt.Println("x now:", x)
}

func main() {
	test()
	fmt.Println("---")
	deferArgs()
}
