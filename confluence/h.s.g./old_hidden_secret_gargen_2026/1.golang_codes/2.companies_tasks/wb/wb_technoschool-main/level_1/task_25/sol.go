package main

import (
	"fmt"
	"time"
)

func sleep(d time.Duration) {
	<-time.After(d)
}

func main() {
	start := time.Now()
	sleep(500 * time.Millisecond)
	fmt.Println(time.Since(start) >= 500*time.Millisecond)
}
