# Паттерн: Double Pointer — The Pointer to Pointer

## 🎯 Суть проблемы

В Go **все передается по значению**. Когда ты передаешь указатель в функцию, передается **копия указателя**, а не сам указатель.

Чтобы изменить **сам указатель** (не объект, на который он указывает), нужен **указатель на указатель** — `**T`.

## 🗡️ Код из задачи

```go
type Person struct {
	Name string
	Age  uint8
}

func changePerson(person **Person) {
	*person = &Person{
		Name: "Vlad",
		Age:  9,
	}
}

func main() {
	person := &Person{
		Name: "Viktor",
		Age:  10,
	}
	
	fmt.Println(person.Name)  // Viktor
	changePerson(&person)      // передаем адрес указателя
	fmt.Println(person.Name)  // Vlad
}
```

**Вывод**:
```
Viktor
Vlad
```

## 🔍 Разбор по слоям памяти

### Шаг 1: Создание объекта

```
┌─────────────────────────┐
│ person (переменная)     │ <-- адрес: 0x1000
│ значение: 0xA000        │ (указатель на Person)
└─────────────────────────┘
              │
              ▼
┌─────────────────────────┐
│ Person в куче           │ <-- адрес: 0xA000
│ Name: "Viktor"          │
│ Age:  10                │
└─────────────────────────┘
```

### Шаг 2: Вызов `changePerson(&person)`

Передаем **адрес переменной `person`** (0x1000):

```go
func changePerson(person **Person)  // person = 0x1000
```

Внутри функции:
- `person` — указатель на указатель (значение: 0x1000)
- `*person` — сам указатель (значение: 0xA000)
- `**person` — объект Person (Name: "Viktor")

### Шаг 3: Присваивание `*person = &Person{...}`

Создается **новый объект** в куче и **перезаписывается указатель**:

```
┌─────────────────────────┐
│ person (переменная)     │ <-- адрес: 0x1000
│ значение: 0xB000        │ (ИЗМЕНЕНО!)
└─────────────────────────┘
              │
              ▼
┌─────────────────────────┐
│ Person в куче (НОВЫЙ)   │ <-- адрес: 0xB000
│ Name: "Vlad"            │
│ Age:  9                 │
└─────────────────────────┘
```

Старый объект (0xA000) остается в памяти, но станет garbage (GC очистит).

## 🧪 Сравнение с одинарным указателем

### ❌ Одинарный указатель: меняет **поля объекта**

```go
func changePersonFields(person *Person) {
	person.Name = "Vlad"  // меняет поле объекта
	person.Age = 9
}

func main() {
	person := &Person{Name: "Viktor", Age: 10}
	changePersonFields(person)
	fmt.Println(person.Name)  // Vlad (объект изменен)
}
```

### ✅ Двойной указатель: меняет **сам указатель**

```go
func changePerson(person **Person) {
	*person = &Person{Name: "Vlad", Age: 9}  // заменяет сам указатель
}

func main() {
	person := &Person{Name: "Viktor", Age: 10}
	changePerson(&person)
	fmt.Println(person.Name)  // Vlad (указатель заменен на новый объект)
}
```

## 💡 Правило Додзё

> **"Одинарный указатель — это карта к дому. Двойной указатель — это карта к месту, где хранится карта к дому. Ты можешь изменить дом (поля) или заменить саму карту (указатель)."**

## 🛠️ Когда использовать `**T`?

### 1. Замена объекта внутри функции

```go
func resetConnection(conn **net.Conn) {
	if *conn != nil {
		(*conn).Close()
	}
	*conn = newConnection()
}
```

### 2. Ленивая инициализация с заменой

```go
func getOrCreate(cache **Cache) *Cache {
	if *cache == nil {
		*cache = NewCache()
	}
	return *cache
}
```

### 3. Реализация двусвязного списка

```go
type Node struct {
	Value int
	Prev  *Node
	Next  *Node
}

func deleteNode(node **Node) {
	if (*node).Prev != nil {
		(*node).Prev.Next = (*node).Next
	}
	if (*node).Next != nil {
		(*node).Next.Prev = (*node).Prev
	}
	*node = nil  // зануляем сам указатель
}
```

## 🧪 Эксперименты

### Тест 1: Копирование указателя

```go
func tryChange(person *Person) {
	person = &Person{Name: "Vlad", Age: 9}  // локальная переменная!
}

func main() {
	person := &Person{Name: "Viktor", Age: 10}
	tryChange(person)
	fmt.Println(person.Name)  // Viktor (не изменилось!)
}
```

**Почему?** Внутри `tryChange` переменная `person` — это **копия** указателя. Присваивание меняет только локальную копию.

### Тест 2: Вывод адресов

```go
func changePerson(person **Person) {
	fmt.Printf("Адрес переменной person в main: %p\n", person)
	fmt.Printf("Значение person (адрес объекта): %p\n", *person)
	*person = &Person{Name: "Vlad", Age: 9}
	fmt.Printf("Новое значение person (адрес нового объекта): %p\n", *person)
}

func main() {
	person := &Person{Name: "Viktor", Age: 10}
	fmt.Printf("Адрес переменной person: %p\n", &person)
	fmt.Printf("Значение person (адрес объекта): %p\n", person)
	changePerson(&person)
}
```

## 🎓 Вопросы для медитации

1. Чем отличается `person = &Person{}` от `*person = Person{}`?
2. Что вернет `person` после `changePerson(&&person)`? (спойлер: не компилируется)
3. Можно ли реализовать `swap(a, b *int)` для обмена значений через двойной указатель?
4. Как работает `flag.StringVar(&myVar, "name", "default", "help")`? Это `**string`?

## 🔗 Связанные паттерны

- **Builder pattern**: часто использует `**T` для замены объектов
- **Factory pattern**: возвращает указатель, но может принимать `**T` для инициализации
- **C interop**: `cgo` часто требует `**C.char` для изменения C-указателей

---

**Задание**: Реализуй функцию `func resetIfNil(p **Person)`, которая создает объект Person только если `*p == nil`. Объясни, почему не сработает `func resetIfNil(p *Person)`.
