# Урок 11 — 20 апреля 2026

Яндекс: структура этапов (скрининг → leetcode → кодинг → SD → финал). В `11_full_task.md` — шаблон конспекта без конкретных условий; ниже — классика скрининга + якоря на company packs.

## Этапы (что говорить / что писать)

| Этап | Фокус | Фразы |
|------|--------|--------|
| Скрининг | лёгкий algo + Go-основы | идея → сложность → edge → тест |
| Leetcode | medium, 20–40 мин | hash/two pointers/stack; не молчи |
| Кодинг | production-кусочек | интерфейсы, ошибки, тесты |
| System design | уточнения → RPS → API → data → scale | «не будь Jimmy» — сначала вопросы |
| Финал | мотивация, легенда, процессы | опыт + trade-offs |

Готовый код Яндекса в gym:
- `old_…/2.companies_tasks/yandex/load_balancer/`
- `old_…/2.companies_tasks/yandex/fin_tech/`

## Задачи (скрининг-разминка)

### 1. Two Sum
Индексы двух чисел с суммой `target`, O(n) через `map[int]int`.

### 2. Valid Parentheses
Стек для `()[]{}`. Краевой: пустая строка → true; лишняя закрывающая → false.

## Как запускать

```bash
export GOWORK=off
cd code/01_two_sum && go run main.go
```
