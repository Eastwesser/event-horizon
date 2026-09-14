package main

import "fmt"

// Классика про общий backing array: y и z делят cap с x после append(x,?).
// При len==cap append аллоцирует новый массив — иначе y и z перетирают один слот.
func main() {
	x := []int{}
	x = append(x, 0)
	x = append(x, 1)
	x = append(x, 2)
	y := append(x, 3)
	z := append(x, 4)
	fmt.Println("y:", y, "z:", z)
	fmt.Printf("x len=%d cap=%d\n", len(x), cap(x))
	// На собесе: «пока cap позволяет, y и z пишут в один и тот же слот за len(x)».
}
