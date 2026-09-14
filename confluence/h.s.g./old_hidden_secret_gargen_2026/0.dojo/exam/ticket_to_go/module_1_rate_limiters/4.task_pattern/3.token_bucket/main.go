package main

// Token Bucket — ведро с токенами (⭐ основной ответ на собесе про API rate limit).
// Альтернатива тикеру: lazy refill в Allow() по elapsed time — без фоновой горутины.
// Burst после простоя: да. Mutex обязателен.

import (
	"fmt"
	"sync"
	"time"
)

/*
Вопрос: что нужно хранить в Token Bucket?

	Текущее количество токенов
	Максимальное количество токенов
	Скорость пополнения (как часто добавляется токен)
	Что ещё для потокобезопасности?
*/
type TokenBucket struct {
	CurrentTokensNumber int
	MaxTokensNumber     int
	RefillRateInterval  time.Duration
	PreviousRefillTime  time.Time
	mu                  sync.Mutex
}

/*
После этого напиши конструктор
NewTokenBucket(maxTokens int, refillRate time.Duration) *TokenBucket.

Что он должен делать?
	Создать структуру
	Инициализировать токены (сколько?)
	Инициализировать время последнего пополнения (когда?)
	Запустить горутину для пополнения токенов (как?)

Использование параметров вместо хардкода:
	Было: CurrentTokensNumber: 10, MaxTokensNumber: 100
	Стало: CurrentTokensNumber: maxTokens, MaxTokensNumber: maxTokens
	Почему: конструктор должен использовать переданные значения, а не фиксированные.

Тип времени:
	Было: PreviousRefillTime: "2026-01-17" (строка)
	Стало: PreviousRefillTime: time.Now() (time.Time)
	Почему: для работы с временем нужен тип time.Time, не строка.

Горутина пополнения:
	Было: обращение к типу TokenBucket вместо экземпляра
	Стало: go tokenBucket.refillTokens() — метод на экземпляре
	Почему: метод вызывается на конкретном экземпляре, не на типе.
*/

func NewTokenBucket(maxTokens int, refillRate time.Duration) *TokenBucket {
	// Создать структуру
	tokenBucket := &TokenBucket{
		CurrentTokensNumber: maxTokens,    // Начальное количество = максимум (ведро полное)
		MaxTokensNumber:     maxTokens,    // Используем переданный параметр
		RefillRateInterval:  refillRate,   // Интервал пополнения
		PreviousRefillTime:  time.Now(),   // Время сейчас (time.Time, не строка!)
		mu:                  sync.Mutex{}, // Инициализация мьютекса
	}

	// Запускаем горутину для пополнения токенов
	go tokenBucket.refillTokens()

	return tokenBucket
}

/*
Метод refillTokens():
	Использует time.Ticker для периодического пополнения
	Блокирует доступ через mu.Lock() для потокобезопасности
	Добавляет токен, если их меньше максимума

Вопрос: почему начальное количество токенов равно максимуму?
(Подсказка: ведро создаётся полным.)
*/

// refillTokens периодически пополняет токены
func (tb *TokenBucket) refillTokens() {
	ticker := time.NewTicker(tb.RefillRateInterval) // Тикер с интервалом пополнения
	defer ticker.Stop()                             // Остановим при завершении

	for range ticker.C { // Каждый тик
		tb.mu.Lock() // Блокируем для потокобезопасности
		if tb.CurrentTokensNumber < tb.MaxTokensNumber {
			tb.CurrentTokensNumber++ // Добавляем один токен
		}
		tb.PreviousRefillTime = time.Now() // Обновляем время последнего пополнения
		tb.mu.Unlock()
	}
}

/*
Теперь нужно написать метод Allow() bool — основной метод Rate Limiter'а.
Вопрос: что должен делать Allow()?
	Проверить, есть ли токены
	Если есть — забрать один токен и вернуть true
	Если нет — вернуть false
	Как это сделать потокобезопасно?
*/

// Allow проверяет, можно ли выполнить запрос (есть ли токен)
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()         // Блокируем для потокобезопасности
	defer tb.mu.Unlock() // Гарантируем разблокировку при выходе

	if tb.CurrentTokensNumber > 0 {
		tb.CurrentTokensNumber-- // Забираем один токен
		return true              // Запрос разрешён
	}

	return false // Токенов нет, запрос отклонён
}

func main() {
	tbRateLimiter := NewTokenBucket(
		5, 
		1 *time.Second,
	) // 5 токенов, пополнение 1/сек

	for i := 0; i < 10; i++ {
		if tbRateLimiter.Allow() {
		    fmt.Printf("Запрос %d: разрешён\n", i)
        } else {
            fmt.Printf("Запрос %d: отклонён\n", i)
        }
		time.Sleep(200 * time.Millisecond)
	}
}
