# Паттерн: Error Wrapping — errors.Is, errors.As, errors.Join

## 🎯 Философия обработки ошибок в Go

В Go ошибка — это **значение**, а не исключение. С Go 1.13 появилась поддержка **wrapping errors** — оборачивания ошибок в цепочки с сохранением контекста.

## 🗡️ Три меча обработки ошибок

### 1. `errors.Is` — проверка на конкретную ошибку

```go
var ErrNotFound = errors.New("not found")

func findUser(id int) error {
	if id < 0 {
		return fmt.Errorf("invalid id: %w", ErrNotFound)
	}
	return nil
}

func main() {
	err := findUser(-1)
	
	if errors.Is(err, ErrNotFound) {
		fmt.Println("User not found")  // Сработает!
	}
}
```

**Суть**: `errors.Is` разворачивает цепочку ошибок и ищет `ErrNotFound` **на любом уровне**.

### 2. `errors.As` — извлечение конкретного типа ошибки

```go
type ValidationError struct {
	Field string
	Msg   string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on %s: %s", e.Field, e.Msg)
}

func validateAge(age int) error {
	if age < 0 {
		return &ValidationError{Field: "age", Msg: "must be positive"}
	}
	return nil
}

func main() {
	err := validateAge(-5)
	
	var vErr *ValidationError
	if errors.As(err, &vErr) {
		fmt.Printf("Field: %s, Message: %s\n", vErr.Field, vErr.Msg)
		// Output: Field: age, Message: must be positive
	}
}
```

**Суть**: `errors.As` ищет в цепочке ошибку **конкретного типа** и извлекает её.

### 3. `errors.Join` — объединение нескольких ошибок

```go
func validateUser(name string, age int) error {
	var errs []error
	
	if name == "" {
		errs = append(errs, errors.New("name is required"))
	}
	if age < 18 {
		errs = append(errs, errors.New("age must be 18+"))
	}
	
	return errors.Join(errs...)  // Go 1.20+
}

func main() {
	err := validateUser("", 15)
	
	if err != nil {
		fmt.Println(err)
		// Output:
		// name is required
		// age must be 18+
	}
}
```

**Суть**: `errors.Join` создает **мультиошибку**, где каждая может быть проверена через `errors.Is` / `errors.As`.

## 🧪 Глубокое сравнение

### `==` vs `errors.Is`

```go
var ErrNotFound = errors.New("not found")

func test() error {
	return fmt.Errorf("user: %w", ErrNotFound)
}

func main() {
	err := test()
	
	// ❌ НЕ РАБОТАЕТ
	if err == ErrNotFound {
		fmt.Println("Not found")  // Не напечатается!
	}
	
	// ✅ РАБОТАЕТ
	if errors.Is(err, ErrNotFound) {
		fmt.Println("Not found")  // Напечатается!
	}
}
```

**Почему?**
- `err == ErrNotFound` сравнивает **адреса** в памяти
- `errors.Is` разворачивает **цепочку** через `Unwrap()`

### Type Assertion vs `errors.As`

```go
type MyError struct{ Msg string }
func (e *MyError) Error() string { return e.Msg }

func test() error {
	return fmt.Errorf("context: %w", &MyError{Msg: "fail"})
}

func main() {
	err := test()
	
	// ❌ НЕ РАБОТАЕТ
	if myErr, ok := err.(*MyError); ok {
		fmt.Println(myErr.Msg)  // Не напечатается!
	}
	
	// ✅ РАБОТАЕТ
	var myErr *MyError
	if errors.As(err, &myErr) {
		fmt.Println(myErr.Msg)  // Напечатается: "fail"
	}
}
```

## 💡 Правило Додзё

> **"Ошибка — это река, текущая через слои кода. `%w` создает поток, `errors.Is` ищет исток, `errors.As` вылавливает рыбу нужного вида."**

## 🛠️ Паттерны использования

### 1. Sentinel Errors (именованные ошибки)

```go
var (
	ErrNotFound    = errors.New("not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrRateLimited = errors.New("rate limited")
)

func getUser(id int) (*User, error) {
	if !isAuthenticated() {
		return nil, fmt.Errorf("user %d: %w", id, ErrUnauthorized)
	}
	// ...
}

// Использование
user, err := getUser(123)
if errors.Is(err, ErrUnauthorized) {
	return errors.New("please login")
}
```

### 2. Custom Error Types

```go
type HTTPError struct {
	Code int
	Msg  string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.Code, e.Msg)
}

func fetchData(url string) error {
	// ...
	return &HTTPError{Code: 404, Msg: "page not found"}
}

// Использование
if err := fetchData("example.com"); err != nil {
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		if httpErr.Code == 404 {
			fmt.Println("Page not found")
		}
	}
}
```

### 3. Error Aggregation

```go
func processFiles(files []string) error {
	var errs []error
	
	for _, f := range files {
		if err := processFile(f); err != nil {
			errs = append(errs, fmt.Errorf("file %s: %w", f, err))
		}
	}
	
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}
```

### 4. Context Wrapping

```go
func saveUser(u *User) error {
	if err := db.Insert(u); err != nil {
		return fmt.Errorf("save user %d: %w", u.ID, err)
	}
	return nil
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	user := parseUser(r)
	
	if err := saveUser(user); err != nil {
		// Цепочка: "save user 123: table users: connection timeout"
		log.Printf("request failed: %v", err)
		
		// Проверка глубинной ошибки
		if errors.Is(err, sql.ErrConnDone) {
			w.WriteHeader(503)
			return
		}
	}
}
```

## 🧪 Эксперименты

### Тест 1: Цепочка из 3 уровней

```go
var ErrDatabase = errors.New("database error")

func level3() error {
	return ErrDatabase
}

func level2() error {
	return fmt.Errorf("query failed: %w", level3())
}

func level1() error {
	return fmt.Errorf("operation failed: %w", level2())
}

func main() {
	err := level1()
	fmt.Println(err)
	// Output: operation failed: query failed: database error
	
	if errors.Is(err, ErrDatabase) {
		fmt.Println("Found ErrDatabase at any level!")  // Сработает!
	}
}
```

### Тест 2: errors.Join с проверкой

```go
func main() {
	err1 := errors.New("error 1")
	err2 := errors.New("error 2")
	
	multi := errors.Join(err1, err2)
	
	fmt.Println(errors.Is(multi, err1))  // true
	fmt.Println(errors.Is(multi, err2))  // true
	
	fmt.Println(multi.Error())
	// Output:
	// error 1
	// error 2
}
```

### Тест 3: Unwrap вручную

```go
func main() {
	err := fmt.Errorf("wrapped: %w", errors.New("original"))
	
	// Разворачивание вручную
	if unwrapped := errors.Unwrap(err); unwrapped != nil {
		fmt.Println(unwrapped)  // "original"
	}
}
```

## 🎓 Вопросы для медитации

1. Чем отличается `fmt.Errorf("error: %v", err)` от `fmt.Errorf("error: %w", err)`?
2. Можно ли использовать `errors.Is` для `nil` ошибки?
3. Что вернет `errors.As(err, &err)`?
4. Как реализовать собственный `Unwrap()` для кастомной ошибки?

## 🔗 Связанные паттерны

- **Retry pattern**: проверка через `errors.Is` для временных ошибок
- **Circuit Breaker**: подсчет ошибок по типу через `errors.As`
- **Error logging**: обогащение контекста через `fmt.Errorf(...%w)`

## 📚 Best Practices

1. **Всегда оборачивай ошибки**: `fmt.Errorf("context: %w", err)`
2. **Используй sentinel errors для публичных API**: `var ErrNotFound = ...`
3. **Custom types для сложных ошибок**: `type ValidationError struct { ... }`
4. **errors.Join для batch операций**: валидация форм, обработка файлов
5. **Не оборачивай дважды**: если контекст уже добавлен, не дублируй

---

**Задание**: Реализуй функцию `func ProcessBatch(items []Item) error`, которая обрабатывает массив элементов и:
1. Возвращает `errors.Join` всех ошибок валидации
2. Немедленно возвращает ошибку при критической ошибке БД
3. Добавляет контекст через `%w` для каждой ошибки

Объясни, почему `errors.Is` предпочтительнее `==` при проверке на критическую ошибку.
