package main

import "fmt"

// distance — кратчайшее расстояние между 'X' и 'Y' в строке из X/Y/O.
func distance(input string) int {
	best := -1
	lastX, lastY := -1, -1
	for i, ch := range input {
		switch ch {
		case 'X':
			lastX = i
			if lastY >= 0 {
				d := lastX - lastY
				if best < 0 || d < best {
					best = d
				}
			}
		case 'Y':
			lastY = i
			if lastX >= 0 {
				d := lastY - lastX
				if best < 0 || d < best {
					best = d
				}
			}
		}
	}
	if best < 0 {
		return 0
	}
	return best
}

func main() {
	fmt.Println(distance("XOOYOX"))
	fmt.Println(distance("OOOO"))
}
