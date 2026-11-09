Что можно улучшить (-):

Нет health checks бэкендов. Если gateway-2 упал, балансер всё равно может выбрать его (least connections = 0, потому что никто не подключается). Запрос уйдёт в мёртвый гейтвей, вернётся 502. Это критично для продакшена. Но для MVP — ок.

Нет retry. Если backend вернул 5xx, балансер не пробует другой. Просто возвращает 502.

lb.mu.RLock() на каждом запросе — это лишний overhead. Список backends не меняется после инициализации, можно было бы не лочить. Но это незначительно (RLock дешёвый).

Логирование в hot path — log.Printf("🔀 Proxying to %s (active: %d)", ...). Это замедляет каждый запрос. Лучше использовать slog с уровнем debug или убрать.

backend.Proxy.ErrorHandler — возвращает 502, но не логирует, какой backend упал. Лучше добавить backend.URL в лог.

Нет graceful degradation — если все backends упали, возвращает 503. Это правильно, но можно добавить кэш (например, отдавать 503 с Retry-After).


---- CHANGES in v.1.1.3.6 ----


Что изменилось:

Backend.Mutex удалён (мёртвое поле).

Healthy atomic.Bool — новое поле.

NewLeastConnBalancer(urls, log) — теперь принимает logger.

getLeastConnBackend пропускает !Healthy и принимает skip для будущего.

ErrorHandler логирует backend.URL.Host.

log.Printf → lb.log.Debug (hot path не пишет в info).

atomic.LoadInt32 в логе (был гонка чтения).

X-Balancer-Backend теперь response header, а не request.

503 → Retry-After: 5.

metrics.ActiveConnections.Set после инкремента (раньше был до выбора, показывал stale).

Backends() и HealthyBackends() — для health-check горутины и /ready.