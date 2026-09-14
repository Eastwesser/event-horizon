package main

import "fmt"

type Statistic struct {
	UserID int
	Steps  int
}

type Result struct {
	UserIDs []int
	Steps   int
}

// getChampions — участники без пропусков дней с макс. суммой шагов (ничья → все).
func getChampions(statistics [][]Statistic) Result {
	maxDays := len(statistics)
	type agg struct{ steps, days int }
	m := make(map[int]agg)

	for day, dayStats := range statistics {
		seen := make(map[int]struct{})
		for _, s := range dayStats {
			if _, dup := seen[s.UserID]; dup {
				continue
			}
			seen[s.UserID] = struct{}{}
			a, ok := m[s.UserID]
			if day == 0 {
				m[s.UserID] = agg{steps: s.Steps, days: 1}
				continue
			}
			if !ok || a.days != day {
				// пропустил предыдущий день — выбыл
				delete(m, s.UserID)
				continue
			}
			a.days++
			a.steps += s.Steps
			m[s.UserID] = a
		}
		// кто не пришёл в этот день — выбывает
		if day > 0 {
			for id, a := range m {
				if a.days != day+1 {
					delete(m, id)
				}
			}
		}
	}

	maxSteps := -1
	var ids []int
	for id, a := range m {
		if a.days != maxDays {
			continue
		}
		if a.steps > maxSteps {
			maxSteps = a.steps
			ids = []int{id}
		} else if a.steps == maxSteps {
			ids = append(ids, id)
		}
	}
	if maxSteps < 0 {
		return Result{}
	}
	return Result{UserIDs: ids, Steps: maxSteps}
}

func main() {
	fmt.Println(getChampions([][]Statistic{
		{{1, 1000}, {2, 1500}},
		{{2, 1000}},
	}))
	fmt.Println(getChampions([][]Statistic{
		{{1, 2000}, {2, 1500}},
		{{2, 4000}, {1, 3500}},
	}))
}
