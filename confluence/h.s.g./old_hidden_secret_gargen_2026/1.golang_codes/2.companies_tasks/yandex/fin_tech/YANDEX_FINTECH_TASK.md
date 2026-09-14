# YANDEX FINTECH TASK

Вы — backend-разработчик в финтех компании.
Компания предоставляет платежные услуги и должна контролировать лимиты пользователей.
Product owner просит создать систему проверки лимитов перед проведением платежей.

## Определения

Платеж:
- id пользователя,
- сумма (в рублях с копейками),
- тип операции (только списание),
- время операции

Лимиты пользователя:
- суточный лимит по сумме (за 24 часа),
- максимальный размер одной операции

История операций:
- список совершенных платежей пользователя

Важно!
- Настройки лимитов пользователей и история платежей предоставляются другими компонентами системы.
- Вам необходимо спроектировать контракты для получения этих данных.
- Реализацию этих компонентов делать не нужно.

## Задача

Написать систему проверки лимитов, которая:
- на вход получает платеж,
- проверяет лимиты,
- возвращает результат проверки: можно ли провести операцию,
- если нельзя, то указывает причину (какой именно лимит будет превышен),
- Проведение платежа не входит в вашу задачу - другая команда займется обработкой платежей после проверки.
- Ваша задача - только проверка возможности проведения платежа.

## Ограничения
В рамках данной задачи считаем, что все платежи одного пользователя происходят строго последовательно. 
Во время проверки лимита не может быть проведен платеж того же пользователя.

```java
class PaymentsChecker {
PaymentsHistoryService paymentsHistoryService;
UserLimitsService userLimitsService;
    public ??? checkPayment(??? payment) { 
        // TODO implement 
    } 
}

interface PaymentsHistoryService {
// TODO any functions
}

interface UserLimitsService {
// TODO any functions
}
```

----------------------------------

## THE CODE

```go
package main

import (
    "context"
    "time"
)

// ========== Модели данных ==========

// Payment — входящий платеж
    type Payment struct {
    UserID    string
    Amount    float64 // в рублях с копейками (можно int64 в копейках, но для простоты float64)
    Timestamp time.Time
}

// CheckResult — результат проверки
    type CheckResult struct {
    Allowed bool
    Reason  string // пустая строка, если Allowed == true
}

// UserLimits — лимиты пользователя
    type UserLimits struct {
    DailyLimit float64 // суточный лимит за 24 часа
    MaxAmount  float64 // максимальная сумма одной операции
}

// ========== Интерфейсы внешних сервисов ==========

// PaymentsHistoryService — сервис истории платежей
type PaymentsHistoryService interface {
    // GetPaymentsSince возвращает все платежи пользователя за указанный период
    GetPaymentsSince(ctx context.Context, userID string, since time.Time) ([]Payment, error)
}

// UserLimitsService — сервис лимитов пользователя
type UserLimitsService interface {
    // GetUserLimits возвращает лимиты пользователя
    GetUserLimits(ctx context.Context, userID string) (*UserLimits, error)
}

// ========== Основная логика ==========

type PaymentsChecker struct {
    historyService PaymentsHistoryService
    limitsService  UserLimitsService
}

func NewPaymentsChecker(history PaymentsHistoryService, limits UserLimitsService) *PaymentsChecker {
    return &PaymentsChecker{
        historyService: history,
        limitsService:  limits,
    }
}

func (c *PaymentsChecker) CheckPayment(ctx context.Context, p Payment) (*CheckResult, error) {
    // 1. Получаем лимиты пользователя
    limits, err := c.limitsService.GetUserLimits(ctx, p.UserID)
    if err != nil {
        return nil, err // можно обернуть, но для простоты так
    }
    if limits == nil {
        return &CheckResult{
            Allowed: false,
            Reason:  "user limits not configured",
        }, nil
    }

	// 2. Проверяем максимальную сумму одной операции
	if p.Amount > limits.MaxAmount {
		return &CheckResult{
			Allowed: false,
			Reason:  "amount exceeds max operation limit",
		}, nil
	}

	// 3. Проверяем суточный лимит (за последние 24 часа)
	since := p.Timestamp.Add(-24 * time.Hour)
	payments, err := c.historyService.GetPaymentsSince(ctx, p.UserID, since)
	if err != nil {
		return nil, err
	}

	// Суммируем все платежи за последние 24 часа
	var sum24h float64
	for _, payment := range payments {
		sum24h += payment.Amount
	}

	// Проверяем, не превысит ли новый платеж суточный лимит
	if sum24h+p.Amount > limits.DailyLimit {
		return &CheckResult{
			Allowed: false,
			Reason:  "daily limit exceeded",
		}, nil
	}

	// Все проверки пройдены
	return &CheckResult{Allowed: true}, nil
}
```

---
## MOCKS

```go
package main

import (
    "context"
    "fmt"
    "time"
)

// Мок-сервисы для примера
type mockHistoryService struct{}

func (m *mockHistoryService) GetPaymentsSince(ctx context.Context, userID string, since time.Time) ([]Payment, error) {
    // В реальном коде тут запрос в БД или другой сервис
    return []Payment{}, nil
}

type mockLimitsService struct{}

func (m *mockLimitsService) GetUserLimits(ctx context.Context, userID string) (*UserLimits, error) {
    return &UserLimits{
        DailyLimit: 1000.00,
        MaxAmount:  500.00,
    }, nil
}

func main() {
    checker := NewPaymentsChecker(
        &mockHistoryService{}, 
        &mockLimitsService{}, 
	)

	payment := Payment{
		UserID:    "user123",
		Amount:    300.00,
		Timestamp: time.Now(),
	}

	result, err := checker.CheckPayment(context.Background(), payment)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if result.Allowed {
		fmt.Println("Payment allowed")
	} else {
		fmt.Printf("Payment denied: %s\n", result.Reason)
	}
}
```

## Важные моменты для Go:

```text
    Что использовал		        Почему

    context.Context		        Для передачи таймаутов, отмены запросов между сервисами
    float64			            Для простоты, но в проде лучше использовать int64 (копейки) или decimal
    error в интерфейсах	        Сервисы могут падать, надо их обрабатывать
    *UserLimits		            nil означает "не найдено", отделяем от ошибок
    Период -24 * time.Hour	    Четко считает последние 24 часа
```

## Если хочешь точность с деньгами (int64 в копейках):

```go
type Payment struct {
    UserID    string
    Amount    int64 // копейки (рубли * 100)
    Timestamp time.Time
}

type UserLimits struct {
    DailyLimit int64 // копейки
    MaxAmount  int64 // копейки
}

// Проверка
if p.Amount > limits.MaxAmount { ... }
if sum24h+p.Amount > limits.DailyLimit { ... }
// Так точнее и нет проблем с float64.
```

## Вопросы, которые могут задать

### 1. Что если история платежей вернула ошибку?
- Сейчас возвращаем `error` → вызывающий сервис решит: retry, fallback или отказать.
- В финтехе обычно `fail closed` (отказать), если не уверены.

### 2. Как учесть, что лимиты могут быть в разных валютах?
- Хороший вопрос. В текущей задаче все в рублях.
- В реальности нужно хранить лимиты с валютой и конвертировать по курсу на момент платежа.

### 3. Что если пользователь делает 1000 платежей в день?
- Сейчас каждый раз запрашиваем все платежи за 24 часа → O(n) по истории.
- Можно хранить агрегированную сумму в Redis и обновлять её.

### 4. Почему не используем мьютексы?
- В условии сказано, что платежи одного пользователя последовательны.
- Значит, на верхнем уровне есть блокировка по `userID`.

### 5. Что с копейками?
- В проде `int64` (копейки) или `decimal.Decimal`.
- `float64` может дать погрешность: `0.1 + 0.2 != 0.3`.
