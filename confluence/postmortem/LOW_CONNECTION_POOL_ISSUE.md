1. SetMaxOpenConns(25) — вот bottleneck
text
services/inventory/internal/app/di.go:140: db, err := sql.Open("postgres", a.cfg.PGDSN())
services/inventory/internal/app/di.go:144: db.SetMaxOpenConns(25)     ← 25!
services/inventory/internal/app/di.go:145: db.SetMaxIdleConns(10)     ← 10
25 соединений на 160 одновременных gRPC запросов = 6.4 на соединение. Каждый SearchItems ждёт свободное соединение ~1.2s. Отсюда логи:

text
SearchItems duration_ms=1171
SearchItems duration_ms=1209
Индекс быстрый (0.25ms), PG не нагружен (count(*) 293 строки, Redis latency 1ms). Но 25 соединений не хватает на 160 параллельных gRPC.

Это классический connection pool starvation. Фикс — увеличить MaxOpenConns до 100 (или больше).