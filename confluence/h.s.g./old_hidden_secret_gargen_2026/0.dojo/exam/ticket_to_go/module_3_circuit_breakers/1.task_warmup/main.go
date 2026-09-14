package main

import "fmt"

/*
Разминка: panic / recover.

- recover ловит панику только в той же горутине и только из defer.
- Код после panic в той же функции не выполняется.
- Паника в другой горутине → процесс падает, если там нет своего recover.
*/
func test() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered:", r)
		}
	}()

	panic("boom!")
	fmt.Println("After panic") // не выполнится
}

func main() {
	test()
	fmt.Println("main continues")

	done := make(chan struct{})
	go func() {
		defer close(done)
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("other goroutine recovered:", r)
			}
		}()
		panic("in other goroutine")
	}()
	<-done
}
