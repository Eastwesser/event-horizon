// Sprint capacity: люди × дни × focus → часы; сравнение с оценкой задач.
package main

import "fmt"

func capacityHours(devs, sprintDays int, focus float64) float64 {
	const hoursPerDay = 8.0
	return float64(devs*sprintDays) * hoursPerDay * focus
}

func main() {
	// 5 разработчиков, спринт 10 раб. дней, focus 0.7 (встречи/ревью)
	cap := capacityHours(5, 10, 0.7)
	tasksSP := 42.0 // «сторипоинты» команды
	hoursPerSP := 4.0
	need := tasksSP * hoursPerSP
	fmt.Printf("capacity=%.0fh need≈%.0fh fit=%v\n", cap, need, need <= cap)
}
