1. P=200 — чистый (после прогрева)
text
                    Прошлый (cold)    Сейчас (warm)    Δ
browse p95          90003ms           10382ms          ×8.7 ✅
submit p95          60211ms           3271ms           ×18.4 ✅
purchase p95        86567ms           15239ms          ×5.7 ✅
browse errors       32.55%            0.00%            ✅
status_5xx          16.09%            0.00%            ✅
circuit shop        открывался        0                ✅
purchase_owned      35.66%            60.79%           (contention)
Cold start был проблемой прошлого прогона. Сейчас — 0% errors, 0% 5xx. circuit shop не открылся. Это — стабильный P=200 на 8 GiB.

2. free -h — вот оно
text
Mem: 9.6Gi total, 6.2Gi used, 322Mi free, 3.6Gi buff/cache, 3.4Gi available
Swap: 4.0Gi total, **1.3Gi used**, 2.7Gi free
Swap: 1.3Gi used — это плохо. Виртуалка свопится. 9.6 GiB RAM, 6.2 GiB used, 3.4 GiB available — впритык. Swap активен — значит, что-то выгружено на диск. Под P=200 это добавляет 10–100ms на каждый page fault.

Смотри docker stats — кто жрёт:

text
clickhouse: 916 MiB    ← 10% RAM
jaeger:     663 MiB    ← 7% RAM
grafana:    321 MiB    ← 3% RAM
prometheus: 150 MiB    ← 1.5% RAM
nats-2:     113 MiB
postgres (auth): 81 MiB
postgres-billing: 95 MiB
postgres-profile: 85 MiB
...
Мониторинг (clickhouse + jaeger + grafana + prometheus + alertmanager + exporters) — ~2.2 GiB. Это 23% RAM на наблюдаемость. Для P=200 на 8 GiB — много.

Плюс — clickhouse 95% CPU, BLOCK I/O 89 GB / 113 GB — ClickHouse активно пишет (analytics ingest из NATS). postgres-billing 2.65% CPU, BLOCK I/O 120 MB / 722 MB — тоже пишет. nats-2 12.83% CPU, BLOCK I/O 75 MB / 2.6 GB — JetStream пишет на диск.

3. Что это значит
8 GiB — впритык для P=200 с полным мониторингом + ClickHouse. Swap 1.3 GiB — сигнал, что RAM не хватает. clickhouse жрёт CPU и диск на ingest из NATS — это фоновый процесс, который мешает P=200.

Это не баг кода. Это потолок 8 GiB all-in-one. README говорит «Prefer 8 vCPU / 32 GiB» для нагрузочных демо. На 8 GiB P=200 — с натяжкой, P=100 — комфортно.