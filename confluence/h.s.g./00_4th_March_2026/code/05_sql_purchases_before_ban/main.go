// In-memory аналог SQL из урока:
// 1) DISTINCT user+SKU для покупок до бана
// 2) пользователи с SUM(price) > 5000 (до бана)
// На собесе: LEFT JOIN ban, WHERE date < date_from OR ban IS NULL; HAVING vs WHERE.
package main

import (
	"fmt"
	"sort"
	"time"
)

type User struct {
	ID, First, Last string
}

type Purchase struct {
	SKU, UserID int
	Price       int
	Date        time.Time
}

type Ban struct {
	UserID   int
	DateFrom time.Time
}

func parse(s string) time.Time {
	t, _ := time.Parse("2006-01-02", s)
	return t
}

func beforeBan(uid int, d time.Time, bans map[int]time.Time) bool {
	from, ok := bans[uid]
	return !ok || d.Before(from)
}

func main() {
	users := []User{
		{"1", "Ivan", "Petrov"},
		{"2", "Anna", "Petrova"},
		{"3", "Anna", "Petrova"},
	}
	purchases := []Purchase{
		{1, 1, 5500, parse("2021-02-15")},
		{1, 1, 5700, parse("2021-01-15")},
		{2, 1, 4000, parse("2021-02-14")},
		{3, 2, 8000, parse("2021-03-01")},
		{4, 2, 400, parse("2021-03-02")},
	}
	bans := map[int]time.Time{1: parse("2021-03-08")}

	type pair struct{ uid, sku int }
	seen := map[pair]struct{}{}
	var uniq []pair
	for _, p := range purchases {
		if !beforeBan(p.UserID, p.Date, bans) {
			continue
		}
		k := pair{p.UserID, p.SKU}
		if _, ok := seen[k]; !ok {
			seen[k] = struct{}{}
			uniq = append(uniq, k)
		}
	}
	sort.Slice(uniq, func(i, j int) bool {
		ui, uj := users[uniq[i].uid-1], users[uniq[j].uid-1]
		if ui.First != uj.First {
			return ui.First < uj.First
		}
		if ui.Last != uj.Last {
			return ui.Last < uj.Last
		}
		return uniq[i].sku < uniq[j].sku
	})
	fmt.Println("DISTINCT user+SKU before ban:")
	for _, p := range uniq {
		u := users[p.uid-1]
		fmt.Printf("  %s %s sku=%d\n", u.First, u.Last, p.sku)
	}

	sum := map[int]int{}
	for _, p := range purchases {
		if beforeBan(p.UserID, p.Date, bans) {
			sum[p.UserID] += p.Price
		}
	}
	fmt.Println("users with total > 5000:")
	for _, u := range users {
		id := 0
		fmt.Sscanf(u.ID, "%d", &id)
		if s := sum[id]; s > 5000 {
			fmt.Printf("  %s | %s | %s | %d\n", u.ID, u.First, u.Last, s)
		}
	}
}
