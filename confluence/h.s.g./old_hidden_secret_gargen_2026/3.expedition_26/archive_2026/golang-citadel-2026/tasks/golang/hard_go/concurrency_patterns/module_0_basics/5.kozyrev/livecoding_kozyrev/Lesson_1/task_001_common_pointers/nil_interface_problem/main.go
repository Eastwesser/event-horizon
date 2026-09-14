// Проблема nil-интерфейса с ненулевым типом
package main

import (
  "errors"
  "fmt"
  "reflect"
)

type MyError struct {
}

func (m *MyError) Error() string {
  return "This is an error"
}

func process() error {
  var m *MyError // zero value for * is nil
  // var m *MyError:
  // Объявляем переменную m как указатель на MyError.
  // Её нулевое значение — это nil (указатель в никуда).

  //if ... {
  //  m = &MyError{}
  //}

  fmt.Println(m == nil) // true
  // Проверяем, равен ли сам указатель nil.
  // Это true, так как мы не инициализировали переменную.

  return m
  /*
    Функция process() возвращает тип error.
    Интерфейс error в Go — это особый тип.
    Интерфейс считается nil только если и его значение, и его тип являются nil.
    Когда мы возвращаем m (указатель, равный nil), компилятор неявно оборачивает его в интерфейс error.
    В результате создается интерфейсная переменная, у которой:
    Тип: *main.MyError (он известен и не является nil!)
    Значение: nil (так как сам указатель m был nil)
    Такой интерфейс уже не является nil.
    Он содержит информацию о типе, но не имеет конкретного значения этого типа.
  */
}

func main() {
  err := errors.New("this is an error") // Нормальная работа с ошибками
  if err != nil {
    fmt.Println(err)
  }

  err1 := process()

  fmt.Println(err1 == nil)          // false
  fmt.Println(reflect.TypeOf(err1)) // *main.MyError
  fmt.Println(reflect.ValueOf(err1).IsNil())
  fmt.Println(reflect.TypeOf(err1) == reflect.TypeOf(true))

  // if the error != nil, we panic
  if err1 != nil {
    panic(err1)
  }
}
