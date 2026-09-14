# QUICK SORT BEST CASE

```go
package main

import (
	"fmt"
)

type User struct {
	ID   int
	Name string
	Age  int
}

func QuickSort(slice []User, less func(a, b User) bool) {
	if len(slice) <= 1 {
		return
	}

	pivotIndex := partition(slice, less)
	QuickSort(slice[:pivotIndex], less)
	QuickSort(slice[pivotIndex+1:], less)
}

func partition(slice []User, less func(a, b User) bool) int {
	pivot := slice[len(slice)-1]
	i := 0

	for j := 0; j < len(slice)-1; j++ {
		if less(slice[j], pivot) {
			slice[i], slice[j] = slice[j], slice[i]
			i++
		}
	}

	slice[i], slice[len(slice)-1] = slice[len(slice)-1], slice[i]
	return i
}

func main() {
	// Тестовые данные
	users := []User{
		{ID: 3, Name: "Charlie", Age: 25},
		{ID: 1, Name: "Alice", Age: 30},
		{ID: 4, Name: "Diana", Age: 20},
		{ID: 2, Name: "Bob", Age: 35},
		{ID: 5, Name: "Eve", Age: 28},
	}

	fmt.Println("🔹 Original slice:")
	for _, user := range users {
		fmt.Printf("ID: %d, Name: %s, Age: %d\n", user.ID, user.Name, user.Age)
	}

	// Сортируем по ID (ascending)
	QuickSort(users, func(a, b User) bool {
		return a.ID < b.ID
	})

	fmt.Println("\n🔹 Sorted by ID (ascending):")
	for _, user := range users {
		fmt.Printf("ID: %d, Name: %s, Age: %d\n", user.ID, user.Name, user.Age)
	}

	// Сортируем по Age (descending)
	QuickSort(users, func(a, b User) bool {
		return a.Age > b.Age // Обратный порядок!
	})

	fmt.Println("\n🔹 Sorted by Age (descending):")
	for _, user := range users {
		fmt.Printf("ID: %d, Name: %s, Age: %d\n", user.ID, user.Name, user.Age)
	}

	// Сортируем по Name (alphabetical)
	QuickSort(users, func(a, b User) bool {
		return a.Name < b.Name
	})

	fmt.Println("\n🔹 Sorted by Name (alphabetical):")
	for _, user := range users {
		fmt.Printf("ID: %d, Name: %s, Age: %d\n", user.ID, user.Name, user.Age)
	}
}
```
