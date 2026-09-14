// Чемпионат по шагам: max сумма среди тех, кто не пропустил ни одного дня.
// Алгоритм: stepsSum + daysPresent → max среди full attendance → collect ties.
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

func getChampions(statistics [][]Statistic) Result {
	stepsSum := make(map[int]int)
	daysPresent := make(map[int]int)
	totalDays := len(statistics)

	for _, day := range statistics {
		today := make(map[int]struct{})
		for _, st := range day {
			stepsSum[st.UserID] += st.Steps
			today[st.UserID] = struct{}{}
		}
		for uid := range today {
			daysPresent[uid]++
		}
	}

	maxSteps := 0
	for uid, sum := range stepsSum {
		if daysPresent[uid] == totalDays && sum > maxSteps {
			maxSteps = sum
		}
	}

	var champs []int
	for uid, sum := range stepsSum {
		if daysPresent[uid] == totalDays && sum == maxSteps {
			champs = append(champs, uid)
		}
	}
	return Result{UserIDs: champs, Steps: maxSteps}
}

func main() {
	ex1 := [][]Statistic{
		{{1, 1000}, {2, 1500}},
		{{2, 1000}},
	}
	fmt.Printf("%+v\n", getChampions(ex1)) // {UserIDs:[2] Steps:2500}

	ex2 := [][]Statistic{
		{{1, 2000}, {2, 1500}},
		{{2, 4000}, {1, 3500}},
	}
	fmt.Printf("%+v\n", getChampions(ex2)) // {UserIDs:[1 2] Steps:5500}
}
