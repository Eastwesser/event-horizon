package main

import "fmt"

var justString string

func createHugeString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = 'a'
	}
	return string(b)
}

func someFunc() {
	v := createHugeString(1 << 10)
	// копируем первые 100 рун, чтобы не удерживать всю большую строку
	r := []rune(v)
	if len(r) > 100 {
		r = r[:100]
	}
	justString = string(r)
}

func main() {
	someFunc()
	fmt.Println(len(justString))
}
