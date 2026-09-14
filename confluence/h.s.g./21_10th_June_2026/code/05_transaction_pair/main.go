package main

import "fmt"

// findTransactionPair — два индекса суммой = target (two sum).
func findTransactionPair(amounts []int, target int) []int {
	seen := make(map[int]int)
	for i, a := range amounts {
		if j, ok := seen[target-a]; ok {
			return []int{j, i}
		}
		seen[a] = i
	}
	return nil
}

func main() {
	fmt.Println(findTransactionPair([]int{100, 500, 250, 750, 1000}, 1000))
}
