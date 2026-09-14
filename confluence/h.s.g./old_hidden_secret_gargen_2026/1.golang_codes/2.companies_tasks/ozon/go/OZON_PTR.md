# OZON: Слайсы и указатели (задача с собеседования)

## Условие

Дан код. Что выведет программа и почему?

```go
package main

import "fmt"

type Stock struct {
    Resource string
    Quantity int
}

func main() {
    stock := []Stock{
        {Resource: "wood", Quantity: 1000},
        {Resource: "stone", Quantity: 2000},
        {Resource: "food", Quantity: 3000},
    }

    stockPtr := &stock[0]

    stockPtr.Quantity += 100 // 1100

    stock = append(stock, Stock{Resource: "water", Quantity: 4000})

    stockPtr.Quantity += 50 // 1150

    fmt.Printf("%d %s\n", stock[0].Quantity, stock[0].Resource)
    fmt.Printf("%d %s\n", stockPtr.Quantity, stockPtr.Resource)
}
```

## Что выведет?

```text
1100 wood
1150 wood
```

## Почему так происходит?

Шаг 1. Инициализация слайса
```go
stock := []Stock{
    {Resource: "wood", Quantity: 1000},
    {Resource: "stone", Quantity: 2000},
    {Resource: "food", Quantity: 3000},
}
```
Создаётся слайс с len=3, cap=3. Underlying array — массив на 3 элемента.

Шаг 2. Указатель на первый элемент
```go
stockPtr := &stock[0]
stockPtr указывает на первый элемент в underlying array.
```

Шаг 3. Первое изменение
```go
stockPtr.Quantity += 100 // 1100
```
Меняем wood в старом массиве → стало 1100.

Шаг 4. Append (критический момент)
```go
stock = append(stock, Stock{Resource: "water", Quantity: 4000})
```
len=3, cap=3 → места нет. Go:

Выделяет новый массив (обычно cap=6)
Копирует все элементы из старого массива в новый
Добавляет water
stock теперь ссылается на новый массив
Старый массив больше не используется.

Шаг 5. Второе изменение
```go
stockPtr.Quantity += 50 // 1150
stockPtr всё ещё указывает на старый массив (на ту же память).
```

Меняем значение в старом массиве → wood = 1150.

Шаг 6. Вывод
```go
fmt.Printf("%d %s\n", stock[0].Quantity, stock[0].Resource)
// stock[0] — из НОВОГО массива: wood = 1100

fmt.Printf("%d %s\n", stockPtr.Quantity, stockPtr.Resource)
// stockPtr — из СТАРОГО массива: wood = 1150
```
Схематично
До append:

```text
старый массив: [wood:1100, stone:2000, food:3000]
stock → ссылается на старый массив
stockPtr → указывает на wood в старом массиве
```
После append:

```text
старый массив: [wood:1150, stone:2000, food:3000]  ← stockPtr всё ещё здесь
новый массив: [wood:1100, stone:2000, food:3000, water:4000]  ← stock теперь здесь
```

Ключевые выводы для собеседования
```text
Правило     	                                                    Почему

Указатели на элементы слайса становятся невалидными после append	При переаллокации создаётся новый массив
Append может вернуть новый слайс с другим underlying array	        Всегда присваивайте результат append обратно
Capacity слайса динамически растёт	                                Обычно удваивается при переаллокации
Не храните указатели на элементы слайса, если планируете append	        Индекс безопаснее
```

Как избежать проблемы?

Вариант 1: Использовать индекс 
```go
stock[0].Quantity += 150
```

Вариант 2: Слайс указателей
```go
stock := []*Stock{
    {Resource: "wood", Quantity: 1000},
}
```

Вариант 3: Заранее выделить capacity
```go
stock := make([]Stock, 0, 10)
stock = append(stock, Stock{...})
```

Что спрашивают дальше?

«А как проверить, произошла ли переаллокация?»

Сравнить cap до и после:

```go
before := cap(stock)
stock = append(stock, newItem)
after := cap(stock)
if after > before {
    // переаллокация была
}
```
