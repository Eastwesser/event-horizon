package main

import (
	"fmt"
	"reflect"
)

type User struct {
	NickName string
}

func main() {
	fmt.Println("Task1: Pointers")

	user := User{
		NickName: "eastwesser",
	}

	fmt.Println(reflect.TypeOf(user), user)

	fmt.Println("Before changing nickname:", user.NickName) // eastwesser
	updateUserInfo(&user) // ⭐️ ПЕРЕДАЕМ УКАЗАТЕЛЬ НА ОРИГИНАЛЬНУЮ СТРУКТУРУ
	fmt.Printf("After changing nickname: %s\n", user.NickName) // Nameless - ТЕПЕРЬ ИМЯ ИЗМЕНИТСЯ!
}

// ⭐️ ПРИНИМАЕМ УКАЗАТЕЛЬ
func updateUserInfo(u *User) {
	u.NickName = "s1ntez"

	fmt.Println("Name in [updateUserInfo]:", u.NickName) // s1ntez
	resetUserInfo(u) // ⭐️ ПЕРЕДАЕМ ТОТ ЖЕ УКАЗАТЕЛЬ (не нужно брать адрес снова)
	fmt.Println("Name after calling [resetUserInfo] inside function [updateUserInfo]:", u.NickName) // Nameless
}

func resetUserInfo(u *User) {
	u.NickName = "rofl" // Изменяет оригинальную структуру
	fmt.Printf("Name after 'rofl' in [resetUserInfo]: %s\n", u.NickName) // rofl
	
	// ⭐️ ЭТО РАБОТАЕТ - заменяем всю структуру по указателю
	*u = User{
		NickName: "Nameless",
	}
	fmt.Printf("Name in [resetUserInfo]: %s\n", u.NickName) // Nameless
}
