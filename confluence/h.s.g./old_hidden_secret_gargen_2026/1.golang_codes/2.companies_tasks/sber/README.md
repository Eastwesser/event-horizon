# Sber — haven pointer

Материалы урока 24:

→ [`../../../../24_unknown_date/`](../../../../24_unknown_date/)  
`24_sber_task.md`, `24_tasks.md`, packs в `code/`.

## Фокус

| Тема | Что говорить | Pack |
|------|--------------|------|
| **Payments** | лимиты отдельно от проведения: daily sum + max single; отказ до мутации | `code/04_payments_checker` |
| **ATM** | greedy по номиналам (RUB/EUR), под mutex; если сдачу не собрать — отказ | `code/05_atm` |
| **Employees filter** | HTTP `?status=`; парсинг query, фильтр слайса/репо | `code/03_employees_filter` |

Дополнительно из того же урока: SQL удаление дублей email (`MIN(id) GROUP BY`), RLE — пересекается с HFLabs.

```bash
cd ../../../../24_unknown_date/code/05_atm
export GOWORK=off
go run main.go
```

На собесе сначала контракты (лимиты / купюры / query), потом код.
