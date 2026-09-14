# Урок 15 — system design / RPS (Сюй)

Теория из голосовух + книга Сюй. Код — калькулятор RPS/QPS для устного SD.

## Что говорить на SD (каркас)

1. **Не будь Jimmy** — сначала уточнения: DAU, read/write, SLA, latency, консистентность.
2. **Масштаб**
   - `RPS_avg = (DAU × actions_per_user_per_day) / 86400`
   - `RPS_peak = RPS_avg × peak_factor` (обычно 2–5, соцсети ~3)
   - `QPS ≈ RPS × queries_per_request` (кэш/БД/граф — отдельно)
3. **Каркас:** клиент → LB → gateway (auth, rate limit) → сервисы → cache → primary/replicas → queue при async.
4. **Stateful sticky sessions** — плохо для scale; state в shared store.
5. **Метрики:** CPU/mem/IO, latency/error rate, DAU/retention/revenue.

Пример из конспекта: 1M DAU × 50 actions = 50M/day → ≈578 RPS avg → ×3 ≈1734 peak.

## Задачи

### 1. RPS calculator
Считает avg/peak RPS и QPS по входным параметрам — чтобы на собесе цифры «сыпались» из головы через формулу.

## Как запускать

```bash
export GOWORK=off
cd code/01_rps_calculator && go run main.go
```
