package main

import "fmt"

/*
Разминка: array vs slice, len/cap, append realocation.

Array — значение фиксированного размера.
Slice — header {ptr, len, cap}; append может сменить backing array.
*/
func main() {
	arr := [5]int{1, 2, 3, 4, 5}
	fmt.Println("array:", arr)

	slice := []int{1, 2, 3}
	fmt.Printf("slice len=%d cap=%d\n", len(slice), cap(slice))

	slice = append(slice, 4)
	fmt.Printf("after append len=%d cap=%d (часто cap вырос)\n", len(slice), cap(slice))

	a := []int{1, 2, 3}
	b := a[:2]
	b = append(b, 99) // может затереть a[2], если cap общий
	fmt.Println("a:", a, "b:", b)
}
