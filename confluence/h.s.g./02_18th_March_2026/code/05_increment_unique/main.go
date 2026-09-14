// IncrementUniqueBy: +delta по уникальным адресам; вернуть slice, updated, duplicates.
package main

import "fmt"

func IncrementUniqueBy(nums []*int, delta int) ([]*int, int, int) {
	seen := make(map[*int]struct{})
	updated, duplicates := 0, 0
	for _, p := range nums {
		if p == nil {
			continue
		}
		if _, ok := seen[p]; ok {
			duplicates++
			continue
		}
		seen[p] = struct{}{}
		*p += delta
		updated++
	}
	return nums, updated, duplicates
}

func main() {
	a, b := 1, 10
	nums := []*int{&a, nil, &b, &a, &a, &b}
	sl, updated, dup := IncrementUniqueBy(nums, 3)
	fmt.Println("a=", a, "b=", b) // 4, 13
	fmt.Print("sl=")
	for _, p := range sl {
		if p == nil {
			fmt.Print("nil ")
			continue
		}
		fmt.Print(*p, " ")
	}
	fmt.Println()
	fmt.Println("updated=", updated, "duplicates=", dup) // 2, 3
}
