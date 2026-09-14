/*
Что выведется и как починить changeName?
Присвоение person = &Person{...} меняет только локальную копию указателя.
Чиним: менять поле person.Name = "Alice" (или **Person / return / метод).
*/
package main

import "fmt"

type Person struct {
	Name string
}

func changeName(person *Person) {
	person.Name = "Alice"
}

func main() {
	person := &Person{Name: "Bob"}
	fmt.Println(person.Name) // Bob
	changeName(person)
	fmt.Println(person.Name) // Alice
}
