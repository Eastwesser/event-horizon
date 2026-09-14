# OHM RESISTOR CHECKER 

```go
/*
    Напиши простую программу Python -> Golang
    
    цвета = ["black", "brown", "red", "orange", "yellow", "green", "blue", "violet", "grey", "white"]
    
    # Получаем числовые значения для каждой цветной полосы
    n = цвета.index((input("Введите 1-й цвет: ")))
    m = цвета.index((input("Введите 2-й цвет: ")))
    p = цвета.index((input("Введите 3-й цвет: ")))
    
    # Рассчитываем сопротивление
    q = int(((n*10) + (m)) * (10**(p)))
    z = q / 1000  # Переводим в килоомы
    
    # Выводим результат
    print("\nЗначение резистора:")
    print(f"{q}Ω и в килоомах: {z}kΩ")
*/

package main

import (
	"fmt"
	"math"
	"strings"
)

func main() {
	colors := []string{
		"black", "brown", "red", "orange", "yellow",
		"green", "blue", "violet", "grey", "white",
	}

	var c1, c2, c3 string

	fmt.Print("1-й цвет: ")
	fmt.Scanln(&c1)
	fmt.Print("2-й цвет: ")
	fmt.Scanln(&c2)
	fmt.Print("3-й цвет (множитель): ")
	fmt.Scanln(&c3)

	idx := func(c string) int {
		c = strings.ToLower(c)
		for i, v := range colors {
			if v == c {
				return i
			}
		}
		return -1
	}

	n, m, p := idx(c1), idx(c2), idx(c3)
	if n == -1 || m == -1 || p == -1 {
		fmt.Println("Ошибка: неверный цвет")
		return
	}

	ohms := (n*10 + m) * int(math.Pow10(p))
	kohms := float64(ohms) / 1000

	fmt.Printf("\nСопротивление: %d Ω = %.2f kΩ\n", ohms, kohms)
}
```
