# Hidden Secret Garden (h.s.g.) — haven for Go interview prep

**Статус: DONE (каркас).** Уроки 00–26 с `README` + `code/`, strikes 20/20, dojo M1–6, printable, company packs углублены. Дальше — **цикл поддерживать / тренировать**, не «строить следующий слой».

**Главная цель:** перед собесом открыл задачу → написал / вспомнил → проговорил стратегию → прошёл. Объёма достаточно для middle+: concurrency, slices/maps, HTTP, dojo (RL/LB/CB/retry/cache), company packs, algo, куски SD.

| Слой | Где | Роль |
|------|-----|------|
| **Code gym** | этот `h.s.g.` | Писать, чинить, прогонять, объяснять вслух |
| **Library** | [`prydwen_knowledge/`](../agents/prydwen_knowledge/README.md) | 1–2 фразы «почему так» |
| **Manuscript** | [`MANUSCRIPT.md`](./MANUSCRIPT.md) | маршрут «собес завтра» |
| **Dojo M1–6** | [`MANUSCRIPT_DOJO.md`](./MANUSCRIPT_DOJO.md) | стратегии ticket_to_go |
| **Уроки 00–26** | [`LESSONS_INDEX.md`](./LESSONS_INDEX.md) | даты, unknown, статусы |
| **Печать** | [`PRINTABLE_INTERVIEW_TASKS.md`](./PRINTABLE_INTERVIEW_TASKS.md) | Go + SQL |
| **Статус** | [`HAVEN_STATUS.md`](./HAVEN_STATUS.md) | что готово / что только руками |

`unrelated/` — **вне скоупа**.

---

## Как пользоваться

1. [`MANUSCRIPT.md`](./MANUSCRIPT.md) — ритуал 90 мин.  
2. Таймер 15–30 мин, без автодополнения.  
3. `export GOWORK=off && go run main.go` (не `go run .`).  
4. Проговорить: идея → сложность → edge → вопрос интервьюера.  
5. Дыра в теории → короткий якорь Prydwen → сразу назад к коду.

Smoke: [`scripts/smoke_mains.sh`](./scripts/smoke_mains.sh).

---

## Карта дерева

```
h.s.g./
├── MANUSCRIPT.md                      ← собес завтра (master)
├── MANUSCRIPT_DOJO.md                 ← M1–M6 детали
├── PRINTABLE_INTERVIEW_TASKS.md
├── LESSONS_INDEX.md / HAVEN_STATUS.md
├── scripts/smoke_mains.sh
├── NN_DATE/                           ← уроки: README + code/*/main.go
│   └── 19_…/denis_sher_tasks/         ← Canon Ката Босса
└── old_hidden_secret_gargen_2026/     ← dojo, strikes, companies
```

| Нужно… | Открывать |
|--------|-----------|
| GMP / Outbox / Kafka словами | `prydwen_knowledge/` |
| Семафор / merge / rate limiter | этот `h.s.g.` |
| Лайвкодинг-поведение | `LIVECODING_NOTES.md` |
| Self-check факапов | `22_unknown_date/22_training_checklist.md` |

---

## Что уже готово

| Слой | Состояние |
|------|-----------|
| Уроки `00`…`26` | у каждой `README.md` + runnable `code/*/main.go` |
| Twenty-strikes | **20/20** |
| Dojo ticket_to_go | **M1–M6** |
| Printable | Go + расширенный SQL |
| Companies | Yandex/Avito runnable; WB TechnoSchool; Ozon/T-Bank packs; Sber/HFLabs в `24_` |
| Denis Sher | primary `19_…/denis_sher_t/` · mirror `old_…/denis_sherb_t/` |

---

## Цикл maintain / train

1. Перед собесом — ритуал из [`MANUSCRIPT.md`](./MANUSCRIPT.md).  
2. Слабое место → повтор + строка в чеклисте `22_…`.  
3. Баг в существующем `main.go` — чинить, не раздувать теорию.  
4. Новые OCR-условия — класть рядом с уроком в `code/`, не плодить третий сад.

Статусы в TODO: `[ ]` нет кода · `[~]` есть, нужно ревью · `[x]` прогнано + объясняю вслух.

---

## Быстрый старт

```bash
export GOWORK=off

cd confluence/h.s.g./19_7th_May_2026/denis_sher_tasks/denis_sher_t/1.semaphore
go run main.go

cd confluence/h.s.g./old_hidden_secret_gargen_2026/0.dojo/exam/ticket_to_go/module_1_rate_limiters/4.task_pattern/3.token_bucket
go run main.go

# smoke уроков + key company packs
./confluence/h.s.g./scripts/smoke_mains.sh
```

---

## Out of scope

- `unrelated/`  
- Подмена HSG чужими курсами  
- OCR скринов за тебя  

---

*HSG = haven. Prydwen = библиотека. MANUSCRIPT.md = «открыл → сказал → показал код».*
