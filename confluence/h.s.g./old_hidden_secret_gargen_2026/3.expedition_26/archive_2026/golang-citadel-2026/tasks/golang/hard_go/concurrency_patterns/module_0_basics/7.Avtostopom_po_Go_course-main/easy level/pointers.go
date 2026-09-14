package main

import "fmt"

// Что выведет код и почему?
// Как изменить v на 3 через функцию без добавления return из функции
func main() {
	v := 5
	p := &v
	fmt.Println(*p)

	changePointer(p)
	fmt.Println(*p)
}

func changePointer(p *int) {
	v := 3
	p = &v
}
