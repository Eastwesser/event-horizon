package main

import "fmt"

func main() {
	fmt.Println(100 + 010)   // что будет выведено? 108
	fmt.Println(100 + 0100)  // что будет выведено? 164
	fmt.Println(100 + 01000) // что будет выведено? 612
}
