# Урок 12 — 25 апреля 2026

Практика: каналы, context (cancel/timeout), синхронные → асинхронные HTTP-запросы со структурой результата.

## Задачи

### 1. Channel R/W
Пишем в канал из горутины, читаем в `main`, закрываем канал после записи. Фраза: «закрывает только отправитель; range до close».

### 2. Context timeout / cancel
Работа прерывается по `ctx.Done()`. На собесе: `WithTimeout` + `defer cancel()`; не клади `context` в структуру.

### 3. Async HTTP status
Параллельные запросы (в демо — локальный `httptest`), результат `{URL, StatusCode, Err}`, `select` на timeout. Не используй «голый» `http.Get` без deadline. Фраза: «буфер канала = число запросов, иначе утечка на send».

Теория из конспекта (устно): GMP, GC, slice/map, pprof; транзакции/индексы/EXPLAIN; Kafka delivery + consumer lag.

## Как запускать

```bash
export GOWORK=off
cd code/01_channel_rw && go run main.go
```
