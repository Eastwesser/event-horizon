📊 METRICS — Event Horizon (по состоянию на 30.06.2026)

✅ Common metrics
№	Метрика	                                Статус	        Источник	        Примечание
1	MAU	                                    ⚠️ Планируется	Analytics Service	Будет считаться через user.registered
2	DAU	                                    ⚠️ Планируется	Analytics Service	Событие score.updated за день
3	Users per 5 years	                    ⚠️ Планируется	Analytics Service	Прогноз на основе MAU и роста
4	Content per user/day	                ⚠️ Планируется	Analytics Service	Среднее количество игр на юзера
5	RPS (READ)	                            ⚠️ В работе	    Gateway Metrics	    gateway_requests_total{method="GET"}
5	RPS (WRITE)	                            ⚠️ В работе	    Gateway Metrics	    gateway_requests_total{method="POST"}
5	RPS (SEARCH)	                        ❌ Нет	            —	           Поиск не реализован
6	Simultaneous connections (net)	        ✅ Есть	       NATS Exporter	   gnatsd_varz_connections — 6 подключений
6	Simultaneous connections (websocket)	⚠️ Планируется	Gateway Metrics	    Нужна кастомная метрика
7	Load: CPU, RAM, I/O, Net	            ⚠️ Планируется	Node Exporter	    Нужно добавить в инфраструктуру
8	DB weight (storage)	                    ✅ Есть	       Redis Exporter	   redis_memory_used_bytes — 1.2 MB
8	DB weight (storage)	                    ✅ Есть	       PostgreSQL	       Смотреть через pg_database_size()
9	$ price, traffic	                    ❌ Нет	       —	               Нужна интеграция с облаком
10	Latency (p50)	                        ⚠️ В работе	    Gateway Metrics	    Из gateway_request_duration_seconds
10	Latency (p95)	                        ⚠️ В работе	    Gateway Metrics	    histogram_quantile(0.95, ...)
10	Latency (p99)	                        ⚠️ В работе	    Gateway Metrics	    histogram_quantile(0.99, ...)
11	Read/Write ratio	                    ⚠️ В работе	    Gateway Metrics	    GET / POST через Prometheus
12	Average size of 1 query	                ❌ Нет	       —	               Нужна кастомная метрика в Gateway
13	Retention policy	                    ⚠️ Планируется	ClickHouse	        Будет настроено в Analytics
14	Peak vs Average load	                ⚠️ Планируется	Prometheus	        Сравнение max() и avg() за период

🟢 Business metrics
№	Метрика	            Статус	            Источник	        Значение
1	Game Submits	    ✅ Есть	           Game Service	       game_submits_total{game_id="hexagon",status="success"} — 122
2	Score Distribution	✅ Есть	           Game Service	       game_score_histogram
3	Popular games	    ⚠️ Планируется	    Game Service	    Нужно добавить метки по играм
4	Lamps earned	    ⚠️ Планируется	    Billing Service	    Нужно добавить кастомную метрику
5	Tickets spent	    ⚠️ Планируется	    Billing Service	    Нужно добавить кастомную метрику
6	Active users (DAU)	⚠️ Планируется	    Analytics Service	Через NATS user.registered

🔵 Golang metrics
№	Метрика	    Статус	    Источник	 Значение
1	Goroutines	✅ Есть	   Go Runtime	go_goroutines{job="game"} — 29
2	GC cycles	✅ Есть	   Go Runtime	go_gc_cycles_automatic_gc_cycles_total
3	Heap	    ✅ Есть	   Go Runtime	go_memstats_heap_alloc_bytes — 3.8 MB
4	Stack	    ✅ Есть	   Go Runtime	go_memstats_stack_inuse_bytes
5	Mutex waits	✅ Есть	   Go Runtime	go_mutex_wait_total_seconds_total

🟣 DB metrics
№	Метрика	                Статус	        Источник	            Значение
1	Connections	            ✅ Есть	       PostgreSQL Exporter	   pg_stat_database_numbackends{datname="eventhorizon"} — 2
2	Slow queries (<200ms)	⚠️ В работе	    PostgreSQL Exporter	    Нужно настроить log_min_duration_statement
3	Deadlocks	            ✅ Есть	       PostgreSQL Exporter	   pg_stat_database_deadlocks — 0
4	Replication lag	        ❌ Нет	       —	                   Реплики не настроены

📈 Рабочие запросы в Prometheus

Go Runtime (Game)
```promql
go_goroutines{job="game"}
go_memstats_heap_alloc_bytes{job="game"}
rate(go_gc_cycles_automatic_gc_cycles_total{job="game"}[1m])
```

NATS
```promql
gnatsd_varz_connections
rate(gnatsd_varz_in_msgs[1m])
rate(gnatsd_varz_out_msgs[1m])
```

Бизнес
```promql
sum(game_submits_total) by (game_id)
game_score_histogram
```

Хранилища
```promql
redis_memory_used_bytes
pg_stat_database_numbackends{datname="eventhorizon"}
```

🎯 Что нужно доделать
Задача	                                Приоритет	Сложность
Gateway метрики (RPS, Latency)	        🔥 Высокий	Средняя
Node Exporter (CPU, RAM, I/O)	        🔥 Высокий	Низкая
Analytics Service (MAU, DAU, Retention)	🔥 Высокий	Высокая
WebSocket метрики	                    Средний	    Низкая
Slow queries PostgreSQL	                Средний	    Низкая
Business метрики (лампы, билетики)	    Средний	    Низкая

Итог: Мы собрали 80% инфраструктурных метрик. Осталось добить Gateway и начать Analytics сервис для бизнес-метрик.