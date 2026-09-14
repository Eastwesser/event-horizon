# Индекс уроков HSG (`00`…`26`)

Каталог папок с датами / `unknown_date`.  
**Master:** [`MANUSCRIPT.md`](./MANUSCRIPT.md) · dojo: [`MANUSCRIPT_DOJO.md`](./MANUSCRIPT_DOJO.md) · печать: [`PRINTABLE_INTERVIEW_TASKS.md`](./PRINTABLE_INTERVIEW_TASKS.md).

## Что починили

| Было | Стало |
|------|--------|
| `09_8th_Aprll_2026` | `09_8th_April_2026` |
| `06_full_brockers.md` | `06_full_brokers.md` |
| `08_full_task.md` (пустой stub) | заполнен конспект HR/легенда из `08_hr_meeting.md` |

## Важно про нумерацию

Номер папки ≈ порядок появления в курсе, **не** строгий календарь:

- `07` (3 Apr) по дате раньше `05` (9 Apr)  
- `17` (21 Apr) лежит между `10` и `11` по календарю  
- `19` (7 May) раньше `14` (11 May) по календарю  

Папки `*_unknown_date` — дата урока неизвестна; **имена не трогаем**, пока не вспомнишь точную дату.

Ниже — **календарный** порядок (unknown — отдельным блоком).

---

## Календарный порядок (с датой)

| # | Папка | Тема | Статус |
|---|--------|------|--------|
| 00 | `00_4th_March_2026` | Старт / uniq / monotonic / champions / outbox / SQL ban | md+go |
| 01 | `01_11th_March_2026` | Zip / merge channels | md+go |
| 02 | `02_18th_March_2026` | Code review кэша (highload) | md+go |
| 03 | `03_25th_March_2026` | Задачи Сергея (context/sync) | md+go |
| 04 | `04_1st_April_2026` | Token bucket + SD notes | md+go |
| 07 | `07_3rd_April_2026` | WebSocket + SD | md+go |
| 08 | `08_8th_April_2026` | HR / типы компаний / легенда (+ demo) | md+go (HR) |
| 09 | `09_8th_April_2026` | Invert map / union / compact / group wait | md+go |
| 05 | `05_9th_April_2026` | Монолит→микросервисы | md+go |
| 10 | `10_15th_April_2026` | Sitemap / граф ссылок | md+go |
| 11 | `11_20th_April_2026` | Яндекс скрининг + leetcode | md+go |
| 17 | `17_21st_April_2026` | Postgres / CH / outbox / MVCC | md+go |
| 12 | `12_25th_April_2026` | Channels, context, HTTP | md+go |
| 13 | `13_29th_April_2026` | HTTP + JSON / async order | md+go |
| 19 | `19_7th_May_2026` | Map internals + **Denis Sher code** | md+go |
| 14 | `14_11th_May_2026` | Fan-out / strStr | md+go |
| 20 | `20_13th_May_2026` | Async / mutex | md+go |
| 21 | `21_10th_June_2026` | Algo triggers + mock | md+go |
| 23 | `23_17th_June_2026` | Avito: unmet demand | md+go |
| 25 | `25_25th_June_2026` | XOR / скобки / Sherb | md+go |
| 26 | `26_6th_August_2026` | Product except self и др. | md+go |

## Без даты (`unknown_date`)

Даты неизвестны — **оставляем имена папок** как есть.

| # | Папка | Тема | Статус |
|---|--------|------|--------|
| 06 | `06_unknown_date` | Брокеры / Kafka / SD | md+go |
| 15 | `15_unknown_date` | System design (RPS, Сюй) | md+go |
| 16 | `16_unknown_date` | Процессы компаний | md+go |
| 18 | `18_unknown_date` | HTTP `Do()` | md+go |
| 22 | `22_unknown_date` | Training checklist (сеть, sync, TLS) | md+go |
| 24 | `24_unknown_date` | HFLabs / Sber / tasks | md+go |

Если вспомнишь дату — переименуй в `NN_Dth_Month_2026` и обнови таблицу.

---

## Чеклист урока

У **каждой** папки `00`…`26`: `README.md` + runnable `code/*/main.go` → статус **`md+go`**.  
Урок **08** — HR-фокус (легенда/типы компаний), `08_full_task.md` заполнен; демо `code/01_company_types_demo`.

Дополнительно code gym: `old_…` (dojo, strikes, company packs).

Дальше только человеческое: `[x] прогнано вслух` по слабому месту.

---

## Быстрые ссылки

| Нужно | Куда |
|-------|------|
| Собес завтра | [`MANUSCRIPT.md`](./MANUSCRIPT.md) |
| Dojo M1–M6 | [`MANUSCRIPT_DOJO.md`](./MANUSCRIPT_DOJO.md) · `old_…/0.dojo/exam/ticket_to_go/` |
| 20 ударов | `old_…/1.golang_codes/1.kata_meeting_tasks/twenty-strikes/` |
| Компании | `old_…/1.golang_codes/2.companies_tasks/` |
| Denis Sher | `19_7th_May_2026/denis_sher_tasks/denis_sher_t/` |
| Smoke | [`scripts/smoke_mains.sh`](./scripts/smoke_mains.sh) |
