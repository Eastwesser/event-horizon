package main

import "fmt"

func deleteAt(s []string, i int) []string {
	copy(s[i:], s[i+1:])
	s[len(s)-1] = "" // avoid memory leak for refs
	return s[:len(s)-1]
}

func main() {
	s := []string{"a", "b", "c", "d", "e"}
	fmt.Println(deleteAt(s, 2)) // remove "c" -> [a b d e]
}
