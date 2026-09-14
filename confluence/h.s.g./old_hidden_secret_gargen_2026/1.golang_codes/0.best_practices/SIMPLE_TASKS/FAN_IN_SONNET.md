# SONNET

```go
/*
Соннет в коде это слияние двух каналов (фан ин)
redPassion
whiteStillness каналы
и нужно реализовать картину, на дженериках.
где мужчина (whiteStillness) и женщина (redPassion) сливаются вместе, это танец (TheirDance)
---

    chan1 (man) -----\
                      v
    chan2 (music) ----> COMBINED chan (dance)
                      ^
    chan3 (woman) ---/

---
======================================================================
*/

package main

import (
	"fmt"
	"sync"
)

// TheirDance — слияние двух (и более) каналов в один танец
func TheirDance[T any](inputChans ...<-chan T) <-chan T {
	var wg sync.WaitGroup
	wg.Add(len(inputChans))

	outputCh := make(chan T)

	// Каждый танцор отдаёт себя общему потоку
	for _, ch := range inputChans {
		go func(ch <-chan T) {
			defer wg.Done()
			for val := range ch {
				outputCh <- val
			}
		}(ch)
	}

	// Когда все закончили — танец окончен, занавес
	go func() {
		wg.Wait()
		close(outputCh)
	}()

	return outputCh
}

func main() {
	whiteStillness := make(chan string)
	redPassion := make(chan string)

	// Оркестр играет
	go func() {
		defer close(whiteStillness)
		defer close(redPassion)

		for i := 0; i < 10; i++ {
			whiteStillness <- fmt.Sprintf("тишина-%d", i)
			redPassion <- fmt.Sprintf("страсть-%d", i)
		}
	}()

	// Танец
	for note := range TheirDance(whiteStillness, redPassion) {
		fmt.Println(note)
	}
}
```
