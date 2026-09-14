[26.03.2026 13:17] Denis Matveev: ’’’go
package main

import (
 "context"
 "errors"
 "fmt"
 "math/rand"
 "sync/atomic"
 "time"
)

/*
Есть приложение с микросервисной архитектурой.
Микросервис можно абстрагировать с помощью интерфейса Backend.
Для доступа к одному экземпляру микросервиса можно использовать
тип BackendImpl, который уже реализован.

Для каждого микросервиса есть несколько десятков запущенных
экземпляров, каждый из которых доступен по своему адресу addr.
Однако отдельные экземпляры микросервиса ненадежны:
они могут падать, быть недоступными либо перегруженными.

Поэтому вам нужно реализовать тип Balancer, который также реализует
интерфейс Backend и осуществляет client-side балансировку нагрузки
между экземплярами микросервиса.
*/

/* ---------------------- Типы и интерфейсы ---------------------- */

// Request / Response — абстракции запросов и ответов, чтобы Backend мог быть любым
type Request interface{}
type Response interface{}

// Backend — интерфейс микросервиса.
// Любой экземпляр Backend (например, BackendImpl) должен уметь обрабатывать Invoke.
type Backend interface {
 Invoke(ctx context.Context, req Request) (Response, error)
}

// SelectStrategy — стратегия выбора backend (Strategy Pattern).
// Это позволяет менять алгоритм выбора backend без переписывания Balancer.
type SelectStrategy interface {
 Choose(backends []*BackendState, visited map[int]struct{}) int
}

/* ---------------------- Реализация конкретного Backend ---------------------- */

// BackendImpl — простой "заглушечный" backend, имитирующий реальный микросервис.
// В реальной системе это был бы RPC/gRPC/HTTP клиент.
type BackendImpl struct {
 addr string
}

// NewBackend создаёт новый BackendImpl для указанного адреса
func NewBackend(addr string) *BackendImpl {
 return &BackendImpl{addr: addr}
}

// Invoke — имитация вызова backend
func (b *BackendImpl) Invoke(ctx context.Context, req Request) (Response, error) {
 // случайная ошибка для демонстрации circuit breaker и retry
 if rand.Float32() < 0.5 {
  return nil, errors.New("backend failed")
 }
 return fmt.Sprintf("response from %s", b.addr), nil
}

/* ---------------------- BackendState ---------------------- */

// BackendState хранит "стейт" конкретного backend
// Это нужно для реализации load balancing и circuit breaker
type BackendState struct {
 backend      Backend   // сам backend
 inFlight     int32     // количество активных запросов (atomic)
 failures     int32     // количество ошибок за recent window (atomic)
 lastFailTime time.Time // время последней ошибки (для circuit breaker)
}

/* ---------------------- Стратегии выбора backend ---------------------- */

// LeastInFlightStrategy — стратегия выбора наименее загруженного backend
type LeastInFlightStrategy struct {
 breakerFail int           // порог количества ошибок для circuit breaker
 breakerWait time.Duration // время ожидания после открытия circuit breaker
}

// Choose — выбирает backend с наименьшей загрузкой, учитывая circuit breaker
func (s *LeastInFlightStrategy) Choose(backends []*BackendState, visited map[int]struct{}) int {
 var candidates []int // список кандидатов с минимальной загрузкой
 minLoad := int32(-1) // минимальная загрузка
 now := time.Now()    // текущее время для circuit breaker

 for i, be := range backends {
  // пропускаем уже посещённые backend в рамках retry
  if _, ok := visited[i]; ok {
   continue
  }

  // Circuit breaker: пропускаем backend, если было слишком много ошибок и мало времени прошло
  if atomic.LoadInt32(&be.failures) >= int32(s.breakerFail) &&
   now.Sub(be.lastFailTime) < s.breakerWait {
   continue
  }

  load := atomic.LoadInt32(&be.inFlight) // текущая загрузка
  if minLoad == -1 || load < minLoad {
   minLoad = load
   candidates = []int{i} // новый минимальный load → новый список кандидатов
  } else if load == minLoad {
   candidates = append(candidates, i) // одинаковая нагрузка → добавляем
  }
 }

 if len(candidates) == 0 {
  return -1 // нет доступных backend
 }

 // jitter: случайный выбор из наименее загруженных → избегаем "горячей точки"
 return candidates[rand.Intn(len(candidates))]
}
[26.03.2026 13:17] Denis Matveev: /* ---------------------- Balancer ---------------------- */

// Balancer — client-side load balancer, реализующий Backend.
// Он умеет:
// - retry при ошибках
// - timeout через context
// - circuit breaker
// - thread-safe подсчёт in-flight и failures без mutex
type Balancer struct {
 backends   []*BackendState
 retryCount int
 reqTimeout time.Duration

 strategy SelectStrategy // стратегия выбора backend (Strategy Pattern)
}

// NewBalancer создаёт балансировщик для списка адресов и стратегии
func NewBalancer(addrs []string, strategy SelectStrategy) *Balancer {
 bs := make([]*BackendState, len(addrs))
 for i, addr := range addrs {
  bs[i] = &BackendState{
   backend: NewBackend(addr),
  }
 }

 // Конфигурация по умолчанию: retry 3 раза, timeout 2 секунды
 return &Balancer{
  backends:   bs,
  retryCount: 3,
  reqTimeout: 2 * time.Second,
  strategy:   strategy,
 }
}

// Invoke — основной метод Balancer
func (b *Balancer) Invoke(ctx context.Context, req Request) (Response, error) {
 n := len(b.backends)
 if n == 0 {
  return nil, errors.New("no backends available")
 }

 visited := make(map[int]struct{}) // чтобы не пробовать один backend дважды в рамках retry

 for attempt := 0; attempt < b.retryCount && len(visited) < n; attempt++ {
  if err := ctx.Err(); err != nil { // проверяем timeout / cancel
   return nil, err
  }

  // выбор backend через стратегию
  idx := b.strategy.Choose(b.backends, visited)
  if idx == -1 {
   break // все backend недоступны
  }

  visited[idx] = struct{}{}
  be := b.backends[idx]

  atomic.AddInt32(&be.inFlight, 1) // увеличиваем счетчик in-flight

  // оборачиваем вызов в timeout context
  ctxc, cancel := context.WithTimeout(ctx, b.reqTimeout)
  res, err := be.backend.Invoke(ctxc, req)
  cancel()

  atomic.AddInt32(&be.inFlight, -1) // уменьшаем счетчик

  if err != nil {
   atomic.AddInt32(&be.failures, 1) // увеличиваем счетчик ошибок
   be.lastFailTime = time.Now()
   continue // пробуем другой backend
  }

  // успешный вызов → сброс ошибок
  atomic.StoreInt32(&be.failures, 0)
  return res, nil
 }

 return nil, errors.New("all backends failed or unavailable")
}

/* ---------------------- Пример использования ---------------------- */

func main() {
 addrs := []string{"a:1", "b:2", "c:3"}

 // создаём стратегию least-in-flight с circuit breaker
 strategy := &LeastInFlightStrategy{
  breakerFail: 3,
  breakerWait: 5 * time.Second,
 }

 b := NewBalancer(addrs, strategy)

 for i := 0; i < 10; i++ {
  res, err := b.Invoke(context.Background(), "request")
  if err != nil {
   fmt.Println("Error:", err)
  } else {
   fmt.Println("Got:", res)
  }
 }
}
’’’