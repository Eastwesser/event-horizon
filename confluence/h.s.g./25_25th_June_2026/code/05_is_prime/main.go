package main

import "fmt"

func isPrime(n int) bool {
	if n <= 1 {
		return false
	}
	for i := 2; i*i <= n; i++ {
		if n%i == 0 {
			return false
		}
	}
	return true
}

func main() {
	for _, n := range []int{1, 2, 17, 18, 97} {
		fmt.Println(n, isPrime(n))
	}
}
