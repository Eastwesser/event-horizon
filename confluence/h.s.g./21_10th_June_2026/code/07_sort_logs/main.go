package main

import "fmt"

// SortLogsBySeverity — Dutch National Flag: 0,1,2 in-place.
func SortLogsBySeverity(logs []int) {
	lo, mid, hi := 0, 0, len(logs)-1
	for mid <= hi {
		switch logs[mid] {
		case 0:
			logs[lo], logs[mid] = logs[mid], logs[lo]
			lo++
			mid++
		case 1:
			mid++
		default:
			logs[mid], logs[hi] = logs[hi], logs[mid]
			hi--
		}
	}
}

func main() {
	logs := []int{1, 0, 2, 1, 0}
	SortLogsBySeverity(logs)
	fmt.Println(logs)
}
