// Full slice expression x[low:high:max]: cap = max-low; append может писать в общий backing.
package main

import "fmt"

func main() {
	x := make([]int, 5, 10)
	for i := range x {
		x[i] = i + 1
	}
	fmt.Println("x", x, "len", len(x), "cap", cap(x))

	y := x[1:3:4] // elements [2,3], len=2, cap=4-1=3
	fmt.Println("y", y, "len", len(y), "cap", cap(y))

	y = append(y, 100) // still within cap → mutates x[3]
	fmt.Println("after append x", x) // [1 2 3 100 5]
	fmt.Println("after append y", y) // [2 3 100]
}
