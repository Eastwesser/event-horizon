package main

import "fmt"

type User struct {
	Name, Surname string
}

// getUniqNameUsers — оставляет первое вхождение каждого имени.
func getUniqNameUsers(users []User) []User {
	seen := make(map[string]struct{})
	var out []User
	for _, u := range users {
		if _, ok := seen[u.Name]; ok {
			continue
		}
		seen[u.Name] = struct{}{}
		out = append(out, u)
	}
	return out
}

func main() {
	users := []User{
		{"Анна", "Петрова"}, {"Анна", "Иванова"}, {"Анна", "Сидорова"},
		{"Мария", "Петрова"}, {"Петр", "Сидоров"}, {"Петр", "Иванов"},
	}
	fmt.Println(getUniqNameUsers(users))
}
