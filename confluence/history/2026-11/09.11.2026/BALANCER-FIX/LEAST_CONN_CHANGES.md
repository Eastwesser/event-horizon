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

