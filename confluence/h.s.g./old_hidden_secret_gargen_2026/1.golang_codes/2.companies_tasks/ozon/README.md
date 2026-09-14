# Ozon — haven map

Два слоя: **теория (playbook)** и **screening (код)**. Не путать.

## Теория / playbook

| Папка | О чём |
|-------|--------|
| `ozon_playbook/` | Obsidian-плейбук: Go, БД, ОС, задачник |
| `go/` | mutex, race, channels, slice, interface, defer, pprof… |
| `db/` | изоляции, локи Postgres, индексы, anomalies |
| `os_network/` | TCP/UDP, HTTP, FD, OOM, syscall, page memory |
| `architecture/` | rate limiter, scaling, proxy, REST vs RPC |
| `data_structures/` | hash table, LRU |

Читать перед теорией на собесе; runnable код — ниже.

## Screening — runnable

| Pack | Условие | Запуск |
|------|---------|--------|
| `code/01_uniq_randn` | [`screening/UNIQ_RANDN.md`](screening/UNIQ_RANDN.md) | `go run main.go` |
| `code/02_is_monotonic` | [`screening/IS_MONOTONIC.md`](screening/IS_MONOTONIC.md) | … |
| `code/03_remove_zeros` | [`screening/REMOVE_ZEROS.md`](screening/REMOVE_ZEROS.md) | … |
| `code/04_zip_two` | [`screening/ZIP.md`](screening/ZIP.md) | … |

```bash
export GOWORK=off
cd code/01_uniq_randn && go run main.go
# …аналогично 02–04
```

Остальные stub'и в `screening/` (slice quirks, SQL join, string) — условия без отдельного pack; готовь устно / через PRINTABLE.

## Cross-link уроков HSG

Пересекается с разборами:
- [`../../../../00_4th_March_2026/`](../../../../00_4th_March_2026/) — `uniq_randn`, `is_monotonic` (+ champions/outbox)
- [`../../../../01_11th_March_2026/`](../../../../01_11th_March_2026/) — `zip_two`, `zip_n`, merge channels

Порядок: screening packs → playbook go/db → урок 00/01 при дырах.
