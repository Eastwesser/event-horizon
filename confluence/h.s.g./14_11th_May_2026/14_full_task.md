package main

import (
	"fmt"
	"sync"
)

func main() {
	var wg sync.WaitGroup
	s1 := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	ch1 := make(chan int, len(s1))
	// выведи значаения из слайса используя паттерн fan-out
    
}


// найди слово в другом слове (подстроку в главной строке)
//func strStr(haystack, needle string) bool {
	
//} 

