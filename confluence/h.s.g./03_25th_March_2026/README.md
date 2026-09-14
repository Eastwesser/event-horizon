# Урок 03 — 25 марта 2026

Client-side балансировщик (Round Robin + retry/timeout) и палиндром по rune. В suggestions — расширенный вариант со Strategy / least-in-flight / circuit breaker.

## Задачи

### 1. Round-robin Balancer
`Balancer` реализует `Backend`: атомарно/под mutex берём next index `% len`, вызываем `Invoke` с timeout context; при ошибке пробуем следующий инстанс (до N попыток). На собесе: не shadowить `b`, `defer cancel` в цикле — cancel сразу; strategy pattern как ДЗ.

### 2. IsPalindrome
Сравнивать `[]rune`, не байты (Unicode). Two pointers left/right. Рекурсия — опционально; на алгосах можно любой язык.

## Как запускать

```bash
export GOWORK=off
cd code/01_round_robin_balancer && go run main.go
```
