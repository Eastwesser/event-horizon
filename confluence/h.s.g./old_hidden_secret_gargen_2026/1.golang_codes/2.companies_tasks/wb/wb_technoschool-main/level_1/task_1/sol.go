package main

import (
	"fmt"
)

// Human представляет человека с базовыми полями и методами.
type Human struct {
	FirstName string
	LastName  string
	Age       int
}

func (h Human) FullName() string {
	if h.LastName == "" {
		return h.FirstName
	}
	return h.FirstName + " " + h.LastName
}

func (h *Human) Birthday() {
	h.Age++
}

func (h *Human) Rename(firstName string, lastName string) {
	h.FirstName = firstName
	h.LastName = lastName
}

// Action встраивает Human
type Action struct {
	Human
	Role string
}

// Do — дополнительный метод только для Action
func (a Action) Do(actionDescription string) string {
	return fmt.Sprintf("%s (%s) выполняет: %s", a.FullName(), a.Role, actionDescription)
}

func main() {
	// создаем Action с встраиваемым Human
	worker := Action{
		Human: Human{FirstName: "Ivan", LastName: "Petrov", Age: 30},
		Role:  "инженер",
	}
	// встроенные методы Human
	fmt.Println("Имя:", worker.FullName())
	worker.Birthday()
	fmt.Println("Возраст после дня рождения:", worker.Age)
	worker.Rename("Petr", "Ivanov")
	fmt.Println("Новое имя:", worker.FullName())

	fmt.Println(worker.Do("пишет код")) // уникальный метод для Action

}
