# Паттерн: Worker Pool — Ограниченный параллелизм

## 🎯 Задача

Обработать **N задач** с использованием **M воркеров** (M < N), чтобы контролировать нагрузку на систему.

## 🗡️ Классическая задача

> Есть 100 URL'ов. Нужно скачать их все, но одновременно может выполняться **не более 5 запросов**.

## 🔍 Код из задачи

```go
const (
	totalURLs        = 100
	concurrencyLimit = 5
)

func download(url int) {
	time.Sleep(time.Duration(100+url%5*100) * time.Millisecond)
	fmt.Printf("Downloaded: %d\n", url)
}

func main() {
	urls := make([]int, totalURLs)
	for i := 0; i < totalURLs; i++ {
		urls[i] = i
	}

	var wg sync.WaitGroup
	ch := make(chan struct{}, concurrencyLimit)  // Семафор

	for _, url := range urls {
		wg.Add(1)
		go func(id int) {
			ch <- struct{}{}  // Захват слота
			defer func() {
				<-ch          // Освобождение слота
				wg.Done()
			}()
			download(id)
		}(url)
	}

	wg.Wait()
}
```

## 🔍 Как работает паттерн?

### 1. Буферизованный канал как семафор

```go
ch := make(chan struct{}, concurrencyLimit)  // Семафор на 5 слотов
```

- **Буфер размера 5** = максимум 5 горутин могут владеть слотом
- `ch <- struct{}{}` — захват слота (блокируется, если буфер полон)
- `<-ch` — освобождение слота

### 2. Механика

```
Время →
URL 0-4:   [██████] 5 активных, остальные ждут
URL 5-9:   [██████] URL 0 завершился → URL 5 стартует
...
URL 95-99: [██████] Последние 5 задач
```

## 🧪 Альтернативные реализации

### Способ 1: Буферизованный канал (текущий)

**Плюсы**:
- Простота
- Все горутины создаются сразу

**Минусы**:
- Создается 100 горутин сразу (overhead)
- Нет переиспользования воркеров

### Способ 2: Пул воркеров с каналом задач

```go
func main() {
	jobs := make(chan int, totalURLs)
	var wg sync.WaitGroup

	// Запускаем 5 воркеров
	for i := 0; i < concurrencyLimit; i++ {
		wg.Add(1)
		go worker(jobs, &wg)
	}

	// Отправляем задачи
	for i := 0; i < totalURLs; i++ {
		jobs <- i
	}
	close(jobs)

	wg.Wait()
}

func worker(jobs <-chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	for url := range jobs {
		download(url)
	}
}
```

**Плюсы**:
- Создается только 5 горутин
- Воркеры переиспользуются
- Более эффективно по памяти

**Минусы**:
- Чуть больше кода

### Способ 3: golang.org/x/sync/semaphore

```go
import "golang.org/x/sync/semaphore"

func main() {
	ctx := context.Background()
	sem := semaphore.NewWeighted(concurrencyLimit)
	var wg sync.WaitGroup

	for i := 0; i < totalURLs; i++ {
		if err := sem.Acquire(ctx, 1); err != nil {
			break
		}
		wg.Add(1)
		go func(url int) {
			defer sem.Release(1)
			defer wg.Done()
			download(url)
		}(i)
	}

	wg.Wait()
}
```

**Плюсы**:
- Стандартная библиотека (x/sync)
- Поддержка context для отмены

### Способ 4: errgroup.Group (с обработкой ошибок)

```go
import "golang.org/x/sync/errgroup"

func main() {
	g := new(errgroup.Group)
	g.SetLimit(concurrencyLimit)  // Go 1.20+

	for i := 0; i < totalURLs; i++ {
		url := i
		g.Go(func() error {
			return downloadWithError(url)
		})
	}

	if err := g.Wait(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}
```

**Плюсы**:
- Встроенная обработка ошибок
- Отмена всех при первой ошибке

## 💡 Правило Додзё

> **"Worker Pool — это мост через реку. Мост имеет 5 дорожек. 100 повозок должны пересечь его. Каждая повозка ждёт, пока дорожка освободится."**

## 🛠️ Когда использовать Worker Pool?

| Задача | Ограничение |
|--------|-------------|
| HTTP-запросы к внешнему API | Rate limit (100 req/s) |
| Запросы к БД | Connection pool (20 conn) |
| CPU-intensive операции | CPU cores (runtime.NumCPU()) |
| Обработка файлов | Открытых дескрипторов (ulimit -n) |
| gRPC-запросы | Connections limit |

## 🎓 Сравнение подходов

| Метод | Горутин | Память | Гибкость | Ошибки |
|-------|---------|--------|----------|--------|
| Семафор | N | Высокая | Низкая | Вручную |
| Worker Pool | M | Низкая | Средняя | Вручную |
| semaphore | N | Средняя | Высокая | Context |
| errgroup | M | Низкая | Высокая | Встроено |

## 🧪 Эксперименты

### Тест 1: Без ограничения

```go
func main() {
	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			fmt.Println(id)
		}(i)
	}
	wg.Wait()
}
```

**Результат**: 1000 горутин одновременно (overhead scheduler'а).

### Тест 2: С ограничением в 10

```go
func main() {
	var wg sync.WaitGroup
	sem := make(chan struct{}, 10)
	
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		sem <- struct{}{}  // Захват
		go func(id int) {
			defer wg.Done()
			defer func() { <-sem }()  // Освобождение
			fmt.Println(id)
		}(i)
	}
	wg.Wait()
}
```

**Результат**: Максимум 10 горутин активны одновременно.

### Тест 3: Worker Pool с таймаутом

```go
func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	jobs := make(chan int)
	var wg sync.WaitGroup

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case job, ok := <-jobs:
					if !ok {
						return
					}
					download(job)
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	go func() {
		for i := 0; i < 100; i++ {
			select {
			case jobs <- i:
			case <-ctx.Done():
				close(jobs)
				return
			}
		}
		close(jobs)
	}()

	wg.Wait()
}
```

## 🎓 Вопросы для медитации

1. Что произойдет, если `ch <- struct{}{}` будет **после** `go func()`?
2. Почему используется `struct{}`, а не `bool` или `int`?
3. Как реализовать динамический Worker Pool (добавление/удаление воркеров)?
4. Чем отличается `semaphore.NewWeighted(5)` от `make(chan struct{}, 5)`?

## 🔗 Связанные паттерны

- **Rate Limiter**: ограничение по времени, а не по количеству
- **Load Balancer**: распределение задач между воркерами
- **Circuit Breaker**: защита воркеров от перегрузки

---

**Задание**: Реализуй Worker Pool, который:
1. Обрабатывает задачи из канала
2. Динамически изменяет количество воркеров через канал `resize chan int`
3. Собирает статистику: сколько задач обработал каждый воркер
4. Завершается по `context.Done()`

Объясни, почему нельзя просто создать 100 горутин для 100 запросов к базе данных.
