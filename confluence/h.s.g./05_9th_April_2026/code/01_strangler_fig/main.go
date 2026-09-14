// Strangler Fig: постепенно переключаем % запросов с monolith на new service.
package main

import (
	"fmt"
	"math/rand/v2"
	"strings"
)

type Service func(req string) string

func monolith(req string) string { return "monolith:" + req }
func modern(req string) string   { return "microservice:" + req }

// route: percentNew ∈ [0,100] — доля на новый сервис.
func route(req string, percentNew int, old, neu Service) string {
	if rand.IntN(100) < percentNew {
		return neu(req)
	}
	return old(req)
}

func main() {
	for _, pct := range []int{0, 20, 50, 100} {
		counts := map[string]int{}
		for i := 0; i < 1000; i++ {
			out := route("order", pct, monolith, modern)
			if strings.HasPrefix(out, "microservice:") {
				counts["new"]++
			} else {
				counts["old"]++
			}
		}
		fmt.Printf("canary %d%% → old=%d new=%d\n", pct, counts["old"], counts["new"])
	}
}
