# Урок 16 — процессы компаний

Устные talking points по Agile/Scrum/ролям. Код опционален: мини-демо capacity спринта.

## Talking points (60–90 сек блоки)

### Методологии
- **Agile** — философия (итерации, люди > процессы).
- **Scrum** — спринты 1–4 нед, роли, артефакты; команды 3–9.
- **Kanban** — WIP-лимит, без спринтов (поддержка/поток).
- **Waterfall** — жёсткие требования, мало изменений.

### Scrum-артефакты
Backlog → Grooming → Sprint (story points) → Daily → Demo → Retro.  
**Post Mortem** — разбор инцидента, без blame.

### Роли (не путать)
| Роль | «За что» |
|------|----------|
| PM (Project) | сроки, бюджет, риски |
| Product Manager | рынок, стратегия «что/зачем» |
| Product Owner | приоритет бэклога (Scrum) |
| BA / SA | бизнес ↔ техспека / API-контракт |
| Architect | структура, интеграции |

### Словарь
Endpoint, API contract, OpenAPI/proto, BFF, MVP.

### Сервисы и команда
~1 разработчик на 2–3 сервиса комфортно. В прод катит DevOps/тимлид после CI + review.  
Оценка: velocity × story points; 40ч ≈ половина двухнедельного спринта.

### Куда расти (личный якорь)
Staff/Principal / Tech Lead без people-management; SD + SRE/perf.

## Задачи

### 1. Sprint capacity (опционально)
Грубый расчёт: люди × дни × focus factor → доступные «человеко-часы» vs оценка задач.

## Как запускать

```bash
export GOWORK=off
cd code/01_sprint_capacity && go run main.go
```
