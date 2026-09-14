// Channel R/W: запись из горутины, чтение range, close отправителем.
package main

import "fmt"

func main() {
	ch := make(chan int)
	go func() {
		defer close(ch)
		for i := 1; i <= 5; i++ {
			ch <- i
		}
	}()
	for v := range ch {
		fmt.Println("got", v)
	}
}
