package main

import (
	"errors"
	"fmt"
	"sync"
)

func quizNilInterface() {
	handler := func(string) error { return errors.New("oi") }
	handler = nil
	check(handler)
}

func check(v any) {
	if v == nil {
		fmt.Println("is nil")
		return
	}
	// typed nil func в any — НЕ == nil
	if fn, ok := v.(func(string) error); ok {
		fmt.Println("is func:", fn == nil)
	}
}

func quizClosure() {
	names := [4]string{"alpha", "beta", "gamma", "delta"}
	var wg sync.WaitGroup
	var k int
	wg.Add(len(names))
	for k = range names {
		go func(k int) {
			defer wg.Done()
			fmt.Println("name:", names[k])
		}(k) // передаём аргументом — иначе гонка/последний k
	}
	wg.Wait()
}

func quizMapStruct() {
	type User struct {
		id   int
		role string
	}
	users := make(map[int]*User)
	users[1] = &User{id: 1, role: "guest"}
	users[1].role = "admin"
	fmt.Println(users[1])
	// map[int]User + users[1].role = ... — не компилируется (не addressable)
}

func main() {
	quizNilInterface()
	quizClosure()
	quizMapStruct()
}
