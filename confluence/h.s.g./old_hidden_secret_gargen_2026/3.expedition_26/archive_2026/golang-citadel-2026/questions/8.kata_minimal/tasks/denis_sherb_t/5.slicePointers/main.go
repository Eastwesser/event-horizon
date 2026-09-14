/*
Что выведется в первом и втором fmt.Println и какие есть способы решения,
чтобы changeName работал как нужно? (всего вроде их как 5)
*/
package main

import "fmt"

type Person struct {
	Name string
}

// 2 вариант - **Person сделать - указатель на указатель
// 3 вариант - вернуть новое значение (дописать *Person) - func changeName(person *Person) *Person {
func changeName(person *Person) {
	// 2 вариант - и тут тоже *person сделать
	person = &Person{
		Name: "Alice", // меняем указатель, но не значение
	}

	// 1 вариант - тупо поменять поле
	// person.Name = "Alice"
}

// 4 вариант: метод на структуре
/*
func (p *Person) changeName() {
	p.Name = "Alice"
}
*/

func main() {
	person := &Person{
		Name: "Bob",
	}

	fmt.Println(person.Name) // Bob
	changeName(person)       // 2 вариант - добавь сюда адрес указателя changeName(&person)
	fmt.Println(person.Name) // Bob - мы получили копию указателя и изменили эту копию
}
