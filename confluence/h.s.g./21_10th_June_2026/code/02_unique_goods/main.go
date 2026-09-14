package main

import "fmt"

// getUniqueGoods — товары, которые есть только у одного продавца.
func getUniqueGoods(sellers map[int][]string) map[int][]string {
	count := make(map[string]int)
	for _, goods := range sellers {
		seen := make(map[string]struct{})
		for _, g := range goods {
			if _, ok := seen[g]; ok {
				continue
			}
			seen[g] = struct{}{}
			count[g]++
		}
	}
	out := make(map[int][]string)
	for id, goods := range sellers {
		seen := make(map[string]struct{})
		for _, g := range goods {
			if _, ok := seen[g]; ok {
				continue
			}
			seen[g] = struct{}{}
			if count[g] == 1 {
				out[id] = append(out[id], g)
			}
		}
	}
	return out
}

func main() {
	sellers := map[int][]string{
		1: {"варежки", "шуба", "валенки", "стол", "шапка", "шарф", "кофта", "рубашка"},
		2: {"шкаф", "тумба", "шапка", "стол", "ложка", "кофта"},
		3: {"вилка", "кофта", "тарелка", "шарф", "шуба", "стол", "кастрюля"},
		4: {"ковер", "тумба", "кофта", "пылесос", "валенки", "ложка", "стол"},
	}
	fmt.Println(getUniqueGoods(sellers))
}
