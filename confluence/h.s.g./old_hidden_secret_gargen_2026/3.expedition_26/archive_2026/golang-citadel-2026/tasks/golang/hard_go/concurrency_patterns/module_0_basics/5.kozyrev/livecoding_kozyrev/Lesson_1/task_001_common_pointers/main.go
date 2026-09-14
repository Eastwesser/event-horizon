// Что выведет эта программа и почему так?
package main

import (
	"fmt"
	"reflect"
)

const TASK_1 string = "Task1: Pointers"

type User struct {
	NickName string
}

// Если не ставим указатель, тогда скопируется вся структура, но лучше ставить, это влияет на работу с копиями
func updateUserInfo(u *User) {
	u.NickName = "s1ntez"

	fmt.Println("Name in [updateUserInfo]:", u.NickName) // s1ntez, поскольку обращаемся к локальному объекту
	resetUserInfo(u)
	//resetUserInfo(&u) // мы передаем адрес локального объекта user, изменения будут так-же применяться наружу
	fmt.Println("Name after calling [resetUserInfo] inside function [updateUserInfo]:", u.NickName) // s1ntez
}

// у контракта reset есть указатель, и теперь копируется уже указатель (а не структура)
func resetUserInfo(u *User) {
	// u.NickName = "rofl" // rofl выведется в 33 строчке вместо s1ntez
	/* 
		Если бы мы написали так, то изменения повлияли бы на объект с 26 строки (u.NickName = "s1ntez"), 
		потому что они указывают на одно и то же местто в памяти 
	*/

	// HOW TO MAKE NAME UPDATABLE? 
	/*
	*u = User{
        NickName: "Nameless",
    } // rofl -> Nameless

	// OR

	u.NickName = "Nameless"
	*/
	
	/* 
		Далее создаем объект - структуру User с новым именем и присваиваем адрес (&) в этот (u) указатель, 
		затирая и забывая ту область памяти, куда указывали изначально.
		Изменения имени никак не отразятся на 26 строке (u.NickName = "s1ntez"), так как это две разные области памяти
	*/
	u = &User{
		NickName: "Nameless",
	} // в локальный указатель с 34-той строки (u *User) - u - мы засовываем новый указатель, теперь указывает на Nameless
	fmt.Printf("Name in [resetUserInfo]: %s\n", u.NickName) // Nameless, можно так (*u) - разыменовать указатель и от него получить объект, на который он указывает, чтобы сделать от него make
	/*
		Имя не обновляется в main() потому что все изменения происходят с локальной копией в updateUserInfo, 
		а переназначение указателя в resetUserInfo не влияет на оригинальные данные.
	*/
}

func main() {
	user := User{
		NickName: "eastwesser",
	}

	fmt.Println(reflect.TypeOf(user), user) // если нужно уточнить тип переменной, и что там в ней лежит

	fmt.Println("Before changing nickname:", user.NickName) // eastwesser, и еще момент, в golang все объекты передаются по значению (копируются)
	updateUserInfo(&user) // (!) тут бы user -> &user 
	fmt.Printf("After changing nickname: %s\n", user.NickName) //  eastwesser, поскольку мы сюда прокинули копию с 25 строки
}

/*
main.User {eastwesser}
Before changing nickname: eastwesser
Name in [updateUserInfo]: s1ntez        ← изменилось на s1ntez
Name in [resetUserInfo]: Nameless       ← локальная замена указателя
Name after calling [resetUserInfo] inside function [updateUserInfo]: s1ntez  ← осталось s1ntez
After changing nickname: s1ntez         ← финальное значение s1ntez

Почему так происходит:

    updateUserInfo(&user) - передали указатель на оригинальный объект ✅

    u.NickName = "s1ntez" в updateUserInfo - изменили оригинальный объект ✅

    resetUserInfo(u) - передали тот же указатель ✅

    u = &User{NickName: "Nameless"} в resetUserInfo - ❌ПРОБЛЕМА


	Что происходит:

    Вы переназначаете локальную переменную-указатель u на новый объект

    Оригинальный объект из main() остается неизменным

    Все дальнейшие операции идут с новым объектом


	Тут либо А, либо В:
	А) Изменяем поле
	func resetUserInfo(u *User) {
		u.NickName = "Nameless"  // Меняем поле существующего объекта
		fmt.Printf("Name in [resetUserInfo]: %s\n", u.NickName) // Nameless
	}

	В) Или замена всего объекта
	func resetUserInfo(u *User) {
    *u = User{  // Заменяем данные по указателю
			NickName: "Nameless",
		}
		fmt.Printf("Name in [resetUserInfo]: %s\n", u.NickName) // Nameless
	}

	Вывод: В Go указатели передаются по значению (копируются адреса), 
	поэтому переназначение локального указателя не влияет на вызывающую функцию. 
	Чтобы изменить оригинальные данные, нужно либо менять поля через указатель, 
	либо заменять данные по указателю с помощью *u = ....
*/
