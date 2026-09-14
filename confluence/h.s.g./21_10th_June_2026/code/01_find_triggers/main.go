package main

import "fmt"

// findTriggers — ручки, где соседние показания отличаются на >= trigger.
func findTriggers(data map[string][]int, trigger int) []string {
	var out []string
	for name, hist := range data {
		for i := 1; i < len(hist); i++ {
			diff := hist[i] - hist[i-1]
			if diff < 0 {
				diff = -diff
			}
			if diff >= trigger {
				out = append(out, name)
				break
			}
		}
	}
	return out
}

func main() {
	in := map[string][]int{
		"web/1/handler": {1, 2, 2, 4, 4, 4, 4},
		"api/6/handler": {0, 0, 0, 0, 1, 3, 0, 1, 0},
		"js/3/handler":  {0, 0, 0, 1, 0, 0, 0, 0, 0},
	}
	fmt.Println(findTriggers(in, 2))
}
