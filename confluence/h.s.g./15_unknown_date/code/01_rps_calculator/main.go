// RPS/QPS calculator — формула Сюй для устного system design.
package main

import "fmt"

const secondsPerDay = 86400.0

func rpsAvg(dau, actionsPerUserPerDay float64) float64 {
	return dau * actionsPerUserPerDay / secondsPerDay
}

func rpsPeak(avg, peakFactor float64) float64 {
	return avg * peakFactor
}

func qps(rps, queriesPerRequest float64) float64 {
	return rps * queriesPerRequest
}

func main() {
	// Пример из конспекта: 1M пользователей × 50 действий
	avg := rpsAvg(1_000_000, 50)
	peak := rpsPeak(avg, 3)
	fmt.Printf("RPS avg=%.0f peak(×3)=%.0f\n", avg, peak) // ~578 / ~1734

	// Твиттер-оценка: 150M × 2 твита
	twAvg := rpsAvg(150_000_000, 2)
	fmt.Printf("twitter-like RPS avg=%.0f peak(×3)=%.0f QPS(×3)=%.0f\n",
		twAvg, rpsPeak(twAvg, 3), qps(rpsPeak(twAvg, 3), 3))

	// «Миллион юзеров, ~100 req/день» — лайфхак из конспекта
	small := rpsAvg(1_000_000, 100)
	fmt.Printf("1M×100/day → avg=%.0f peak(×5)=%.0f\n", small, rpsPeak(small, 5))
}
