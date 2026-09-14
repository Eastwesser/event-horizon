# Урок 26 — product except self + crawl + gotchas

## Тема
Prefix/suffix products без деления, crawler с лимитом k, битовые маски, quiz «что выведет».

## Задачи (`code/`)

| Pack | Что |
|------|-----|
| `01_product_except_self` | left/right passes |
| `02_crawl_semaphore` | ≤k concurrent GET |
| `03_bitmasks` | `|=` и `&^=` |
| `04_quiz_gotchas` | nil in any, closure, map*User |

## Как решать / что говорить
- **Product:** два прохода, без division, O(n) время / O(1) доп. (кроме результата).
- **Crawl:** семафор `chan struct{}` ёмкости k + WaitGroup.
- **Map gotcha:** элемент map не addressable → храни `*User` или read-modify-write целиком.

## Собес-советы
- Quiz 5 (send на closed chan) — паника; в демо не воспроизводим, проговори устно.
- WithdrawBalance из md — устно: `UPDATE … WHERE balance>=$2 RETURNING …` + tx.

## Как запускать

Путь содержит `h.s.g.` (точки) — `go` может подхватить `go.work` монорепо и сломаться. Всегда:

```bash
cd code/NN_name
export GOWORK=off
go run main.go
```

