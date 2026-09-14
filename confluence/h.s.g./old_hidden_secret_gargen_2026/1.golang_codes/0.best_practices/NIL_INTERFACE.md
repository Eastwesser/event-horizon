# Паттерн: Nil Interface — The Silent Trap

## 🎯 Суть проблемы

В Go интерфейс состоит из **двух частей**:
1. **Type** (тип конкретной реализации)
2. **Value** (значение этой реализации)

Интерфейс считается `nil` **только если обе части nil**.

## 🗡️ Код из задачи

```go
type A interface {
	M()
}
type someStruct struct {}
func (s someStruct) M() {}

func newA() A {
	var s *someStruct  // s = (*someStruct, nil)
	return s           // A = (*someStruct, nil) — НЕ nil!
}

func newB() A {
	return nil         // A = (nil, nil) — это nil
}

func main() {
	fmt.Println(newA() == newB())  // false
}
```

**Вывод**: `false`

## 🔍 Почему false?

| Функция | Type | Value | `== nil`? |
|---------|------|-------|-----------|
| `newA()` | `*someStruct` | `nil` | **false** |
| `newB()` | `nil` | `nil` | **true** |

`newA()` возвращает **typed nil** — интерфейс знает тип `*someStruct`, хоть значение и `nil`.

## 🧪 Эксперименты

### Тест 1: Проверка на nil
```go
func main() {
	a := newA()
	fmt.Println(a == nil)  // false
	
	// Но вызов метода паникует!
	a.M()  // panic: nil pointer dereference
}
```

### Тест 2: Type Assertion
```go
func main() {
	a := newA()
	
	// Type assertion покажет тип
	if s, ok := a.(*someStruct); ok {
		fmt.Printf("Type: %T, Value: %v, Is nil: %v\n", s, s, s == nil)
		// Output: Type: *someStruct, Value: <nil>, Is nil: true
	}
}
```

### Тест 3: Правильная проверка
```go
func isReallyNil(a A) bool {
	if a == nil {
		return true
	}
	
	// Проверка через reflection
	v := reflect.ValueOf(a)
	return v.Kind() == reflect.Ptr && v.IsNil()
}

func main() {
	fmt.Println(isReallyNil(newA()))  // true
	fmt.Println(isReallyNil(newB()))  // true
}
```

## 💡 Правило Додзё

> **"Интерфейс — это шкатулка (type) с мечом (value). Пустая шкатулка (nil, nil) отличается от шкатулки с надписью 'меч', но без самого меча (*T, nil)."**

## 🛠️ Как избежать?

```go
// ❌ BAD: возвращает typed nil
func newA() A {
	var s *someStruct
	return s
}

// ✅ GOOD: явно возвращаем nil
func newA() A {
	var s *someStruct
	if s == nil {
		return nil
	}
	return s
}

// ✅ BEST: не допускаем nil внутри
func newA() A {
	return &someStruct{}
}
```

## 🎓 Вопросы для медитации

1. Почему `var i interface{} = (*int)(nil)` не равен `nil`?
2. Как работает `if err != nil` с интерфейсом `error`?
3. Что вернет `fmt.Printf("%v", newA())`?
4. Может ли `nil` удовлетворять интерфейс с методами?

## 🔗 Связанные паттерны

- Error handling: `return nil, err` vs `return nil, (*MyError)(nil)`
- Empty interfaces: `interface{}` vs `any` (Go 1.18+)
- Type switches с nil значениями

---

**Задание**: Напиши функцию `SafeReturn() A`, которая возвращает `nil`, но при этом имеет логику создания `*someStruct`. Объясни, почему твоё решение безопасно.

// ========== NIL INTERFACE TRAP ==========

```go
func demonstrateNilInterfaceTrap() {
    fmt.Println("\n=== Nil Interface Trap ===")

    var repo UserRepository
    fmt.Printf("repo == nil: %v (type: %T, value: %v)\n", repo == nil, repo, repo)
    
    // Опасность: присваивание nil pointer
    var pgRepo *PostgresUserRepository = nil
    repo = pgRepo  // repo теперь НЕ nil!
    
    fmt.Printf("repo == nil: %v (type: %T, value: %v)\n", repo == nil, repo, repo)
    fmt.Println("Почему? Interface содержит (type=*PostgresUserRepository, value=nil)")
    
    // Правильная проверка
    if repo == nil || repo.(*PostgresUserRepository) == nil {
        fmt.Println("Repository is truly nil")
    }
}
```
