# MANUSCRIPT — маршрут «собес завтра»

Открыл → **сказал** → показал код. Теория-якорь: `../agents/prydwen_knowledge/`. Код-зал: этот `h.s.g./`.

Dojo M1–M6 подробно: [`MANUSCRIPT_DOJO.md`](./MANUSCRIPT_DOJO.md).  
Статус haven: [`HAVEN_STATUS.md`](./HAVEN_STATUS.md).

Запуск везде: `export GOWORK=off && go run main.go` (не `go run .` — в папках с точками ломается module path).

---

## Миссия

Собрать уже лежащий в HSG код в один маршрут middle+ Golang: перед любым собесом за 90 минут пройти удары → один dojo-паттерн → pack целевой компании → устный чеклист факапов → и уметь произнести короткие фразы без воды. Не читать курс заново — тренировать **своё** дерево.

---

## Ритуал «собес завтра» (90 мин)

| # | Блок | Мин | Что делать |
|---|------|-----|------------|
| 1 | **Strikes** | 40 | `twenty-strikes` 1–10 без подсказок, `GOWORK=off`, `go run main.go` |
| 2 | **Dojo** | 20 | Один модуль M1–M6 вслух (стратегия → edge → показать код) — см. [`MANUSCRIPT_DOJO.md`](./MANUSCRIPT_DOJO.md) |
| 3 | **Company** | 20 | 15-мин маршрут целевой фирмы (ниже) + 5 мин запас на edge |
| 4 | **Oral checklist** | 7 | Факап-фразы TCP/UDP · TLS · WaitGroup · Mutex · Once · close канала |
| 5 | **Fail phrases** | 3 | Ещё раз вслух 2–3 слабых пункта из `22_…/22_training_checklist.md` |

Таймер жёсткий. IDE-автодополнение выключить.

---

## Карта haven

| Слой | Путь | Роль |
|------|------|------|
| **Печать** | [`PRINTABLE_INTERVIEW_TASKS.md`](./PRINTABLE_INTERVIEW_TASKS.md) | Go + SQL на любой middle+ |
| **Уроки 00–26** | [`LESSONS_INDEX.md`](./LESSONS_INDEX.md) | даты, unknown, статусы `md+go` |
| **Dojo M1–M6** | [`MANUSCRIPT_DOJO.md`](./MANUSCRIPT_DOJO.md) + `old_…/0.dojo/exam/ticket_to_go/` | RL / LB / CB / health / retry / LRU |
| **Twenty-strikes** | `old_hidden_secret_gargen_2026/1.golang_codes/1.kata_meeting_tasks/twenty-strikes/` | 20/20 runnable |
| **Companies** | `old_…/1.golang_codes/2.companies_tasks/` · индекс `README.md` | Yandex, Avito, T-Bank, WB, Ozon, … |
| **Denis Sher CANON** | `19_7th_May_2026/denis_sher_tasks/denis_sher_t/` | primary |
| **Denis Sher mirror** | `old_…/1.golang_codes/1.kata_meeting_tasks/denis_sherb_t/` | дубль |
| **Факапы / чеклист** | `22_unknown_date/22_training_checklist.md` | устные провалы |
| **Sber / HFLabs** | `24_unknown_date/` | company extract + `code/` |
| **Smoke** | [`scripts/smoke_mains.sh`](./scripts/smoke_mains.sh) | прогон `main.go` уроков + key packs |

---

## Факап-фразы (что сказать)

Источник: `22_unknown_date/22_training_checklist.md`. Коротко, без воды.

### TCP vs UDP
> TCP — соединение, ACK, порядок: HTTP/SSH/БД; критичные логи — TCP + батчинг. UDP — без handshake, быстрее, потери ок: DNS, VoIP, стриминг, игры.

### TLS handshake
> ClientHello → ServerHello + сертификат → клиент проверяет CA/срок/домен → обмен ключами (DH/RSA) → симметричное шифрование на сессии.

### WaitGroup
> `Add` до старта горутины, в горутине `defer Done`, снаружи `Wait`. Забыл `Done` — вечный hang; `Add` внутри горутины без синхрона — race.

### Mutex
> `Lock`/`Unlock` вокруг критической секции. RWMutex — много читателей, один писатель. Без мьютекса на shared map — data race; держать lock на долгом I/O нельзя.

### Once
> `once.Do(fn)` — `fn` ровно один раз на процесс (ленивый синглтон / init). Повторные вызовы — no-op.

### Закрытие каналов
> Закрывает **писатель**. Из закрытого: `v, ok := <-ch` → `ok==false`. `range` сам выходит после close. Закрыть дважды / закрыть читателем — panic.

---

## Per-company 15-мин маршруты

Корень packs: `old_hidden_secret_gargen_2026/1.golang_codes/2.companies_tasks/`.  
Относительные пути от `h.s.g./`.

### Yandex (~15)
1. `old_…/2.companies_tasks/yandex/load_balancer/` — least-conn / unhealthy (5 мин код + фраза).  
2. `old_…/2.companies_tasks/yandex/fin_tech/` — лимиты (5 мин).  
3. Урок-якорь: `11_20th_April_2026/` (two sum / parentheses) + dojo M2 (5 мин вслух).

### Avito (~15)
1. `old_…/2.companies_tasks/avito/champions/` + `CHAMPIONS.md` (8 мин).  
2. `old_…/2.companies_tasks/avito/INTERVIEW_PHASES.md` — фазы (2 мин).  
3. Урок: `23_17th_June_2026/` unmet demand (5 мин стратегия).

### T-Bank (~15)
1. `old_…/2.companies_tasks/t-bank/tasks.md` — взять 1 граф/конкурентность (10 мин руками на бумаге → `main.go` если есть `t-bank/code/`).  
2. Strikes 1–5 или Denis Sher `3.mergeChans` (5 мин) — банк любит sync/каналы.

### WB (~15)
1. `old_…/2.companies_tasks/wb/wb_top10_questions.txt` — 3 вопроса вслух (5 мин).  
2. Один `wb/wb_technoschool-main/level_3/task_N/` с `main.go` / `sol.go` (10 мин).

### Ozon (~15)
1. `old_…/2.companies_tasks/ozon/ozon_playbook/` — 1 тема screening (5 мин).  
2. `ozon/os_network/` или `ozon/go/` — один вопрос из md (5 мин).  
3. Dojo M1 Token Bucket + факап TCP/TLS (5 мин) — у Ozon часто сеть/Go internals.

### Sber / HFLabs (~15)
1. `24_unknown_date/24_sber_task.md` **или** `24_hflabs_task.md` — условие (3 мин).  
2. Один pack из `24_unknown_date/code/` (`04_payments_checker` / `05_atm` / `07_lru_cache`) (10 мин).  
3. Фраза LRU / платежные лимиты из README урока (2 мин).

---

## Canon: Denis Sher

| | Путь |
|--|------|
| **Primary** | `19_7th_May_2026/denis_sher_tasks/denis_sher_t/` |
| **Mirror** | `old_hidden_secret_gargen_2026/1.golang_codes/1.kata_meeting_tasks/denis_sherb_t/` |

Задачи: `1.semaphore` · `2.mutateSlice` · `3.mergeChans` · `4.sliceFilter` · `5.slicePointers` · `6.whatsWrong`.

```bash
export GOWORK=off
cd 19_7th_May_2026/denis_sher_tasks/denis_sher_t/1.semaphore
go run main.go
```

---

## Сквозной ответ middle+ (30 сек)

> На входе — **rate limit**. Между инстансами — **load balancer**. На sync-вызовах наружу — **timeouts + retry с backoff/jitter**, снаружи **circuit breaker** (open → 503). Живость — **health/ready**. Чтение — **cache-aside** (как EH Inventory).

---

*Master manuscript. Dojo-детали — в MANUSCRIPT_DOJO.md.*
