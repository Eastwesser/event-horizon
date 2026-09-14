# Урок 04 — 1 апреля 2026

Token bucket rate limiter по IP, code-review HTTP/SQL сниппета, full slice expression / append и shared backing array. Плюс SD: мессенджер Avito (теория в md).

## Задачи

### 1. Token bucket
На IP — бакет: refill `rate * elapsed`, cap `maxTokens`, `CanAccept` списывает 1. Mutex; cleanup неактивных. На собесе: lazy refill без тикера; float64 токены.

### 2. Code-review bugs
JSON tags без backticks, loop variable capture, SQL injection, read after Close body, context.TODO, секреты в DSN. Демо печатает список замечаний + безопасный паттерн.

### 3. Slice `a[low:high:max]`
`y := x[1:3:4]` → len 2 cap 3? wait: high-low=2, max-low=3... In the notes: y := x[1:3:4] → [2,3] len=2 cap=3? They said cap=4 which would be wrong - max-low = 4-1 = 3. Looking at notes again: `y := x[1:3:4] // [2,3],3,4 cap = 4` - they said cap=4 but mathematically max-low=3. Actually in Go: `s[low:high:max]` capacity is `max-low`. So cap=3. Then append 100 shares backing → x becomes [1,2,3,100,5]. Wait if cap is 3, append of one element when len=2 uses existing capacity. Yes. Their comment said cap=4 which is a mistake in the notes - I'll use correct Go semantics and mention it.

## Как запускать

```bash
export GOWORK=off
cd code/01_token_bucket && go run main.go
```
