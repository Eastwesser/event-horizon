## Разбор задачи с собеседования AvitoTech✅

**Чемпионы по шагам**

```go
package main

import (
    "fmt"
    "reflect"
)

type Statistics struct {
    UserId int
    Steps  int
}

type Result struct {
    UserIds []int
    Steps   int
}

func getChampions(statistics [][]Statistics) Result{}
```

Условие:
Нужно найти чемпионов соревнований по шагам и они должны соответствовать данным критериям:
1. Прошли наибольшее общее количество шагов за все дни соревнования.
2. Не пропустили ни одного дня (то есть у пользователя есть запись в каждом дне).

Входные данные:
- statistics: список дней соревнования
- Каждый день — это список словарей вида: {userId: int, steps: int}

Выходные данные:
- champions: объект вида {userIds: []int, steps: int}

Пояснение:
- userIds — список победителей (если несколько с одинаковым максимумом)
- steps — общее количество шагов победителя(ей)

Пример 1:
statistics = [
    [ { userId: 1, steps: 1000 }, { userId: 2, steps: 1500 } ],
    [ { userId: 2, steps: 1000 } ]
]

Вывод:
champions = { userIds: [2], steps: 2500 }

Пример 2:
statistics = [
    [ { userId: 1, steps: 2000 }, { userId: 2, steps: 1500 } ],
    [ { userId: 2, steps: 4000 }, { userId: 1, steps: 3500 } ]
]

Вывод:
champions = { userIds: [1, 2], steps: 5500 }

```go
package main

import (
    "fmt"
    "reflect"
)

type Statistics struct {
    UserId int
    Steps  int
}

type Result struct {
    UserIds []int
    Steps   int
}

func getChampions(statistics [][]Statistics) Result {
    if len(statistics) == 0 {
    return Result{}
}


// Подсчёт общего кол-ва дней, в которых участвовал каждый пользователь
daysParticipated := make(map[int]int)

// Подсчёт общего кол-ва шагов каждого пользователя
totalSteps := make(map[int]int)

totalDays := len(statistics) // Общее кол-во дней, за которые считались шаги

// Алгоритмическая сложность O(n*m)
for _, day := range statistics { // Выделяем один из дней, пример: [ { userId: 1, steps: 1000 }, { userId: 2, steps: 1500 } ]
    for _, stat := range day {
        daysParticipated[stat.UserId]++
        totalSteps[stat.UserId] += stat.Steps
    }
}

// Валидируем пользователей и находим тех, которые не пропустили ни одного дня
// Алгоритмическая сложность O(m)
validUsers := make(map[int]bool)
    for userId, days := range daysParticipated {
        if days == totalDays {
        validUsers[userId] = true
    }
}

// Находим максимальное кол-во шагов среди провалидированных пользователей
maxSteps := 0
// Алгоритмическая сложность O(v)
for userId, _ := range validUsers {
    if totalSteps[userId] > maxSteps {
        maxSteps = totalSteps[userId]
    }
}

// Собираем всех пользователей с maxSteps
var resultIds []int
// Алгоритмическая сложность O(v)
for userId, _ := range validUsers {
    if totalSteps[userId] == maxSteps {
        resultIds = append(resultIds, userId)
    }
}

return Result{
    UserIds: resultIds,
    Steps:   maxSteps,
}

// Общая алгоритмическая сложность O(n * m) + O(m) + O(v) + O(v)
// n - count days
// m - count users
// v - count usersValid
}

func main() {
    example1 := getChampions([][]Statistics{
        {{UserId: 1, Steps: 1000}, {UserId: 2, Steps: 1500}},
        {{UserId: 2, Steps: 1000}},
    })
    fmt.Println("Example 1 passed:", reflect.DeepEqual(example1, Result{UserIds: []int{2}, Steps: 2500}))
    
    example2 := getChampions([][]Statistics{
        {{UserId: 1, Steps: 2000}, {UserId: 2, Steps: 1500}},
        {{UserId: 2, Steps: 4000}, {UserId: 1, Steps: 3500}},
    })
    fmt.Println("Example 2 passed:", reflect.DeepEqual(example2, Result{UserIds: []int{1, 2}, Steps: 5500}))
    
    example3 := getChampions([][]Statistics{
        {{UserId: 1, Steps: 1000}, {UserId: 2, Steps: 1500}},
        {},
        {{UserId: 2, Steps: 1000}},
    })
    fmt.Println("Example 3 test, случай если один из дней будет пропущен всеми пользователями", example3)
}
```

## Условия:
Найти пользователей с максимальной суммой шагов за все дни
Учитывать только тех, кто участвовал во всех днях (нет пропусков)
Если несколько пользователей имеют одинаковый максимум — вернуть всех

## Критика текущего решения:
- Решение корректное, но есть моменты для улучшения:
- Лишняя мапа validUsers — можно сразу проверять валидность
- Два прохода для поиска максимума и сбора результатов — можно за один
- Потенциальная проблема с пустыми днями — в примере 3 пустой день должен дисквалифицировать всех

## Улучшенное решение:

```go
package main

import (
    "fmt"
)

type Statistics struct {
    UserId int
    Steps  int
}

type Result struct {
    UserIds []int
    Steps   int
}

func getChampions(statistics [][]Statistics) Result {
    if len(statistics) == 0 {
    return Result{}
}

	// Мапы для подсчета дней и шагов
	daysCount := make(map[int]int)
	totalSteps := make(map[int]int)
	totalDays := len(statistics)

	// Собираем статистику за все дни
	for _, day := range statistics {
		// Учитываем только непустые дни
		if len(day) == 0 {
			// Если день пустой, никто не участвовал -> все дисквалифицированы
			return Result{}
		}
		
		// Уникальные пользователи в этом дне (на случай дубликатов)
		seenToday := make(map[int]bool)
		
		for _, stat := range day {
			// Защита от дубликатов в одном дне
			if !seenToday[stat.UserId] {
				daysCount[stat.UserId]++
				seenToday[stat.UserId] = true
			}
			totalSteps[stat.UserId] += stat.Steps
		}
	}

	// Находим чемпионов за один проход
	maxSteps := 0
	var champions []int

	for userId, days := range daysCount {
		// Проверяем участие во всех днях
		if days != totalDays {
			continue
		}
		
		steps := totalSteps[userId]
		
		// Если нашли нового лидера
		if steps > maxSteps {
			maxSteps = steps
			champions = []int{userId}
		} else if steps == maxSteps && steps > 0 {
			// Если разделили первое место
			champions = append(champions, userId)
		}
	}

	return Result{
		UserIds: champions,
		Steps:   maxSteps,
	}
}

func main() {
// Тест 1: Базовый случай
example1 := getChampions([][]Statistics{
{{UserId: 1, Steps: 1000}, {UserId: 2, Steps: 1500}},
{{UserId: 2, Steps: 1000}},
})
fmt.Printf("Пример 1: %+v (ожидается [2] 2500)\n", example1)

	// Тест 2: Несколько победителей
	example2 := getChampions([][]Statistics{
		{{UserId: 1, Steps: 2000}, {UserId: 2, Steps: 1500}},
		{{UserId: 2, Steps: 4000}, {UserId: 1, Steps: 3500}},
	})
	fmt.Printf("Пример 2: %+v (ожидается [1 2] 5500)\n", example2)

	// Тест 3: Пустой день дисквалифицирует всех
	example3 := getChampions([][]Statistics{
		{{UserId: 1, Steps: 1000}, {UserId: 2, Steps: 1500}},
		{}, // Пустой день
		{{UserId: 2, Steps: 1000}},
	})
	fmt.Printf("Пример 3: %+v (ожидается [] 0)\n", example3)

	// Тест 4: Никто не участвовал во всех днях
	example4 := getChampions([][]Statistics{
		{{UserId: 1, Steps: 1000}},
		{{UserId: 2, Steps: 2000}},
		{{UserId: 3, Steps: 3000}},
	})
	fmt.Printf("Пример 4: %+v (ожидается [] 0)\n", example4)

	// Тест 5: Дубликаты в одном дне
	example5 := getChampions([][]Statistics{
		{{UserId: 1, Steps: 1000}, {UserId: 1, Steps: 500}}, // Дубликат
		{{UserId: 1, Steps: 2000}},
	})
	fmt.Printf("Пример 5: %+v (ожидается [1] 3500)\n", example5)
}
```

## Оптимизированная версия с ранним выходом

```go
/*
    func getChampionsOptimized(statistics [][]Statistics) Result {
        if len(statistics) == 0 {
        return Result{}
    }
*/
func getChampionsOptimized(statistics [][]Statistics) Result {
	totalDays := len(statistics)
	
	// Множество пользователей, участвовавших в первом дне
	// Все остальные уже не могут быть чемпионами
	potentialChampions := make(map[int]bool)
	daysCount := make(map[int]int)
	totalSteps := make(map[int]int)
	
	// Обрабатываем первый день отдельно
	if len(statistics[0]) == 0 {
		return Result{} // Пустой первый день
	}
	
	for _, stat := range statistics[0] {
		potentialChampions[stat.UserId] = true
		daysCount[stat.UserId] = 1
		totalSteps[stat.UserId] = stat.Steps
	}
	
	// Обрабатываем остальные дни
	for dayIdx := 1; dayIdx < totalDays; dayIdx++ {
		day := statistics[dayIdx]
		if len(day) == 0 {
			return Result{} // Пустой день
		}
		
		seenToday := make(map[int]bool)
		
		// Обновляем только потенциальных чемпионов
		for _, stat := range day {
			if potentialChampions[stat.UserId] {
				if !seenToday[stat.UserId] {
					daysCount[stat.UserId]++
					seenToday[stat.UserId] = true
				}
				totalSteps[stat.UserId] += stat.Steps
			}
		}
		
		// Удаляем тех, кто не участвовал в этом дне
		for userId := range potentialChampions {
			if !seenToday[userId] {
				delete(potentialChampions, userId)
			}
		}
		
		// Если не осталось потенциальных чемпионов
		if len(potentialChampions) == 0 {
			return Result{}
		}
	}
	
	// Находим победителей среди оставшихся
	maxSteps := 0
	var champions []int
	
	for userId := range potentialChampions {
		if daysCount[userId] == totalDays {
			steps := totalSteps[userId]
			if steps > maxSteps {
				maxSteps = steps
				champions = []int{userId}
			} else if steps == maxSteps {
				champions = append(champions, userId)
			}
		}
	}
	
	return Result{UserIds: champions, Steps: maxSteps}
}
```

## Ключевые моменты решения:

Обработка edge cases:
- Пустые дни (дисквалифицируют всех)
- Дубликаты в одном дне
- Пользователи, участвовавшие не во всех днях

Оптимизации:
- Ранний выход при пустом дне
- Учет только "потенциальных чемпионов"
- Один проход для поиска максимума и сбора результатов

Сложность:
- Время: O(N*M) где N — дни, M — пользователи в день
- Память: O(K) где K — уникальные пользователи

Дополнительные улучшения:
- Можно использовать sync.Map для параллельной обработки дней
- Можно добавить валидацию входных данных
- Можно использовать sort для упорядочивания результатов
- Задача хорошо проверяет понимание структур данных, работу с мапами и edge cases в Go!