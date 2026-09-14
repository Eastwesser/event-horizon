package main

import (
	"fmt"
	"sort"
)

/*
AvitoTech Champions: максимум шагов за все дни + присутствие каждый день.
*/
type Statistics struct {
	UserId int
	Steps  int
}

type Result struct {
	UserIds []int
	Steps   int
}

func getChampions(statistics [][]Statistics) Result {
	if len(statistics) == 0 {
		return Result{}
	}
	days := len(statistics)
	total := map[int]int{}
	present := map[int]int{}
	for _, day := range statistics {
		seen := map[int]bool{}
		for _, s := range day {
			total[s.UserId] += s.Steps
			if !seen[s.UserId] {
				present[s.UserId]++
				seen[s.UserId] = true
			}
		}
	}
	best := -1
	var ids []int
	for uid, sum := range total {
		if present[uid] != days {
			continue
		}
		if sum > best {
			best = sum
			ids = []int{uid}
		} else if sum == best {
			ids = append(ids, uid)
		}
	}
	sort.Ints(ids)
	if best < 0 {
		return Result{}
	}
	return Result{UserIds: ids, Steps: best}
}

func main() {
	fmt.Println(getChampions([][]Statistics{
		{{1, 1000}, {2, 1500}},
		{{2, 1000}},
	})) // { [2] 2500 }
	fmt.Println(getChampions([][]Statistics{
		{{1, 2000}, {2, 1500}},
		{{2, 4000}, {1, 3500}},
	})) // { [1 2] 5500 }
}
