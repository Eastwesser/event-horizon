# Урок 25 — XOR / brackets / intersect / parallel URLs

## Тема
Классика Ката + параллельный опрос URL с сохранением порядка. Сырьё: `25_denis_sherb_tasks.md`, `25_full_task.md`.

## Задачи (`code/`)

| Pack | Что |
|------|-----|
| `01_xor_unique` | single number через XOR |
| `02_valid_brackets` | стек скобок |
| `03_intersect` | мультимножество пересечения |
| `04_remove_duplicates` | in-place unique sorted |
| `05_is_prime` | trial до √n |
| `06_parallel_urls` | WaitGroup + слот по индексу |

## Как решать / что говорить
- XOR: `a^a=0`, `a^0=a` → остаётся уникальный.
- Parallel URLs: `resp := make([]T, n)` + запись в `resp[i]` — порядок гарантирован без канала.
- Не забудь `Timeout` на client и `Body.Close()`.

## Собес-советы
- В черновиках md был двойной `process` и missing import sync — в паке починено.
- Для intersect строй частотную мапу по меньшему массиву.

## Как запускать

Путь содержит `h.s.g.` (точки) — `go` может подхватить `go.work` монорепо и сломаться. Всегда:

```bash
cd code/NN_name
export GOWORK=off
go run main.go
```

