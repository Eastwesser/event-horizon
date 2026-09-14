/*
Что выведется и как починить changeName?

Было: оба Println печатают "Bob".
Почему: person внутри changeName — копия указателя. Присвоение person = &Person{...}
меняет только локальную копию, не объект по адресу и не переменную в main.

Способы починить (на собесе перечислить 4–5):
1) Менять поле: person.Name = "Alice"
2) Двойной указатель: func changeName(pp **Person) { *pp = &Person{Name: "Alice"} }
3) Возвращать новое значение: func changeName(p *Person) *Person { return &Person{Name: "Alice"} }
4) Метод на *Person: func (p *Person) ChangeName() { p.Name = "Alice" }
5) Передать &person если в main лежит Person по значению, и менять через указатель поле
*/
package main

import "fmt"

type Person struct {
	Name string
}

func changeName(person *Person) {
	person.Name = "Alice" // вариант 1 — правим объект по адресу
}

func main() {
	person := &Person{Name: "Bob"}

	fmt.Println(person.Name) // Bob
	changeName(person)
	fmt.Println(person.Name) // Alice
}
