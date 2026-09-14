# Decorator: Инсайты от Мусаси

## 🔑 Суть паттерна

**Decorator (декоратор для каналов)** — оборачивает канал и преобразует данные на лету.

**Принцип:** Канал типа `T` → канал типа `T` (или `U`) с преобразованными данными

**Аналогия:** Декоратор в паттернах проектирования — добавляет поведение, не меняя интерфейс.

## 📊 Когда использовать?

- **Преобразование данных:** `chan int` → `chan string` (форматирование)
- **Валидация/фильтрация:** проверка данных перед передачей
- **Адаптация интерфейса:** API принимает канал, но нужен другой формат данных
- **Декоратор для потоков:** добавить логирование, метрики, трансформацию

## 🏗️ Архитектура

```
Входной канал (T) ──> Decorate ──> Выходной канал (T или U)
                      (декоратор)
```

**Важно:** Мы не можем изменить данные внутри входного канала напрямую (он только для чтения), поэтому создаём новый канал.

## ⚙️ Реализация

### Базовый вариант (тот же тип)

```go
func Decorate[T any](inputCh <-chan T, action func(T) T) <-chan T {
    outputCh := make(chan T)
    
    go func() {
        defer close(outputCh)
        for value := range inputCh {
            outputCh <- action(value)  // Применяем функцию преобразования
        }
    }()
    
    return outputCh
}
```

**Использование:**
```go
inputCh := make(chan int)
outputCh := Decorate(inputCh, func(x int) int { return x * 2 })
```

### Расширенный вариант (разные типы)

```go
func Decorate[T, U any](inputCh <-chan T, fn func(T) U) <-chan U {
    outputCh := make(chan U)
    
    go func() {
        defer close(outputCh)
        for value := range inputCh {
            outputCh <- fn(value)  // Преобразуем T → U
        }
    }()
    
    return outputCh
}
```

**Использование:**
```go
inputCh := make(chan int)
outputCh := Decorate(inputCh, func(x int) string { 
    return fmt.Sprintf("Number: %d", x) 
})
```

## 🔄 Decorator vs Map (функциональное программирование)

**Decorator — это Map для каналов:**

- **Map в списках:** `map([]int, fn) → []string`
- **Decorator в каналах:** `Decorate(chan int, fn) → chan string`

**Принцип тот же:** применяем функцию к каждому элементу.

## 🎭 Декоратор vs Адаптер

### Декоратор (твой случай)
- **Добавляет поведение:** преобразование данных
- **Не меняет интерфейс:** вход и выход — каналы
- **Пример:** `Decorate(chan int, multiply) → chan int`

### Адаптер
- **Меняет интерфейс:** другой тип данных
- **Пример:** `Decorate(chan int, toString) → chan string`

**В Go:** Decorator может быть и декоратором, и адаптером (зависит от функции преобразования).

## ⚠️ Ловушки

### 1. Незакрытый канал

**Проблема:**
```go
go func() {
    for i := 0; i < 5; i++ {
        channel <- i
    }
    // Забыли close(channel) → Decorate будет ждать бесконечно
}()
```

**Решение:**
```go
go func() {
    defer close(channel)  // Всегда закрываем после записи
    for i := 0; i < 5; i++ {
        channel <- i
    }
}()
```

### 2. Паника в функции преобразования

**Проблема:** Если `action` паникует, горутина упадёт, канал не закроется.

**Решение:** Обработка паники (опционально):
```go
go func() {
    defer close(outputCh)
    defer func() {
        if r := recover(); r != nil {
            // Логируем ошибку, но не паникуем
        }
    }()
    for value := range inputCh {
        outputCh <- action(value)
    }
}()
```

### 3. Буферизованные каналы

**Вопрос:** Нужен ли буфер для `outputCh`?

**Ответ:** Зависит от скорости обработки:
- Если преобразование быстрое → можно без буфера
- Если преобразование медленное → буфер защитит от блокировки

## 💡 Вопросы для размышления

1. Можно ли использовать Decorator для фильтрации? (Да, если функция возвращает `nil` или используем отдельный паттерн Filter)
2. В чём разница между Decorator и Pipeline? (Pipeline — это цепочка Decorators)
3. Можно ли сделать Decorator с контекстом для отмены? (Да, добавить `ctx context.Context`)

## 🎯 Связь с другими паттернами

- **Pipeline** — цепочка Decorators: `Decorate → Decorate → Decorate`
- **Filter** — частный случай Decorator (возвращает только подходящие значения)
- **Fan-Out + Decorator + Fan-In** — распределение → преобразование → сбор

## 🔧 Практические примеры

### Пример 1: Форматирование чисел
```go
numbers := make(chan int)
formatted := Decorate(numbers, func(n int) string {
    return fmt.Sprintf("Value: %d", n)
})
```

### Пример 2: Валидация данных
```go
users := make(chan User)
validUsers := Decorate(users, func(u User) *User {
    if u.Age < 18 {
        return nil  // Фильтруем несовершеннолетних
    }
    return &u
})
```

### Пример 3: Адаптация для API
```go
// API принимает chan string, но у нас chan int
numbers := make(chan int)
strings := Decorate(numbers, func(n int) string {
    return strconv.Itoa(n)
})
api.Process(strings)  // API получает нужный тип
```

---

*"Преобразовать — это не просто изменить форму. Это понять, как одно становится другим, сохраняя суть потока."* — Мусаси
