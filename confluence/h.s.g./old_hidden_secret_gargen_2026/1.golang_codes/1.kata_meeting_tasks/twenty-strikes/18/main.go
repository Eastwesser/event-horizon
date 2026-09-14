package main
import ("fmt"; "sort")
// Strike 18: Avito-style — sum |need - closest product|
func totalUnmet(products, needs []int) int {
	sort.Ints(products)
	sum := 0
	for _, need := range needs {
		i := sort.SearchInts(products, need)
		best := 0
		if i == 0 { best = products[0]
		} else if i == len(products) { best = products[len(products)-1]
		} else {
			a, b := products[i-1], products[i]
			if need-a <= b-need { best = a } else { best = b }
		}
		if best > need { sum += best - need } else { sum += need - best }
	}
	return sum
}
func main() {
	fmt.Println(totalUnmet([]int{1, 3, 10}, []int{2, 5, 9}))
}
