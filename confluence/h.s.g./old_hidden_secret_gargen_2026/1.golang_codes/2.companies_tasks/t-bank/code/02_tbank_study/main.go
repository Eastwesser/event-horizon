// minReplacements: мин. замен, чтобы s содержала подстроки "tbank" и "study".
// На собесе: перебор стартов обоих паттернов; при overlap — конфликт символов.
package main

import "fmt"

const (
	patA = "tbank"
	patB = "study"
	plen = 5
)

func minReplacements(s string) int {
	n := len(s)
	if n < plen {
		return -1 // по условию |s|≥10, но на демо страхуемся
	}
	const inf = int(1e9)
	best := inf

	costPair := func(i, j int) int {
		// need[p] = требуемый символ; 0 = свободен
		need := make([]byte, n)
		mark := func(start int, pat string) bool {
			for k := 0; k < plen; k++ {
				p := start + k
				c := pat[k]
				if need[p] != 0 && need[p] != c {
					return false // конфликт overlap
				}
				need[p] = c
			}
			return true
		}
		if !mark(i, patA) || !mark(j, patB) {
			return inf
		}
		cost := 0
		for p := 0; p < n; p++ {
			if need[p] != 0 && s[p] != need[p] {
				cost++
			}
		}
		return cost
	}

	for i := 0; i+plen <= n; i++ {
		for j := 0; j+plen <= n; j++ {
			if c := costPair(i, j); c < best {
				best = c
			}
		}
	}
	if best == inf {
		return -1
	}
	return best
}

func main() {
	demos := []struct {
		s        string
		expected int
	}{
		{"tbankstudy", 0},   // уже обе подстроки
		{"studytbank", 0},   // обе, другой порядок
		{"xxxxxxxxxx", 10},  // две непересекающиеся пятёрки → 10 замен
		{"tbankxxxxx", 5},   // tbank есть, study с нуля
		{"tttttttttt", 8},   // overlap возможен; brute найдёт минимум
	}
	for _, d := range demos {
		got := minReplacements(d.s)
		fmt.Printf("%q → %d (expected %d)\n", d.s, got, d.expected)
	}
}
