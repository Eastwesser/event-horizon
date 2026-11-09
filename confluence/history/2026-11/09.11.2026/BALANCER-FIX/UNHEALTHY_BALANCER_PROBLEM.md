Хорошая новость: балансер работает (curl :8079/ready → 200, X-Balancer-Backend в response, метрики balancer_backend_healthy{...} 1 на всех трёх). Плохая новость: healthcheck в compose не проходит — (unhealthy) на обоих. Разберём почему.

🔴 Почему (unhealthy)
Healthcheck в compose:

yaml
test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://127.0.0.1:8079/ready"]
wget в alpine есть (ты пересобрал образ — [2/3] CACHED RUN apk add --no-cache ca-certificates wget). Но unhealthy — значит wget не получает 200.

Смотри, что проверяет healthcheck: http://127.0.0.1:8079/ready — это /ready балансера, а не /metrics-сервера (:9098). Но /ready есть на :9098, а не на :8079!

Смотри app.go:

go
addr := getenv("BALANCER_ADDR", ":8079")
a.httpServer = &http.Server{Addr: addr, Handler: a.lb}      // ← :8079 — только балансировщик

metricsAddr := getenv("METRICS_ADDR", ":9098")
a.metricsSrv = &http.Server{Addr: metricsAddr, Handler: metricsHandler(a.lb)}  // ← :9098 — /health, /ready, /metrics
/ready живёт на :9098. На :8079 — только ServeHTTP балансера, который проксирует всё в gateway. http://127.0.0.1:8079/ready → балансер выбирает gateway → gateway получает GET /ready → gateway отвечает 200 (у него есть /ready)... стоп, тогда почему unhealthy?

Потому что балансер на :8079 ищет healthy gateway. Если gateway не поднят (а в момент make deploy они ещё стартуют), getLeastConnBackend возвращает nil → 503 No available backends. Или gateway отвечает 200 — тогда wget должен пройти.

Но ты видишь unhealthy на обоих балансерах. Значит либо gateway не отвечает, либо wget не может достучаться до :8079. Скорее первое — на старте make deploy gateway'и ещё не готовы.

Но даже если gateway готов — healthcheck через :8079/ready семантически неправильный: он проверяет «проксирует ли балансер в живой gateway», а не «жив ли сам балансер». Правильный healthcheck — на :9098/ready, который проверяет HealthyBackends() > 0 (ты это уже сделал).

Плюс есть вторая проблема: в metrics.go /ready возвращает 503, если HealthyBackends() == 0. На старте, пока health-check горутина не сделала первый check(), Healthy = true (optimistic в NewLeastConnBalancer). Но если gateway реально не готов — первый check() через 5s → Healthy = false → /ready → 503. Это и есть unhealthy.

--

✅ Фикс — healthcheck на :9098/ready
Файл: deployments/docker-compose.cluster.yml

Место: секции balancer и balancer-2, healthcheck.test.

Сниппет (заменить в обоих):

yaml
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://127.0.0.1:9098/ready"]
      interval: 10s
      timeout: 3s
      retries: 5
      start_period: 10s
Почему :9098: /ready (осмысленный, с HealthyBackends()) живёт именно там. На :8079 — только прокси.

Почему start_period: 10s: дать gateway'ям подняться. Если make deploy поднимает gateway до балансера (depends_on: - gateway), 10s хватит.

НО — тут есть ещё одна проблема, которую я вижу из make deploy:

text
✘ Container deployments-balancer-2  Error dependency balancer-2 failed to start  23.1s
Compose поднимает nginx, у которого depends_on: balancer: condition: service_healthy. Балансер unhealthy → nginx не стартует → make deploy падает.

Это замкнутый круг:

make deploy поднимает все сервисы ($(COMPOSE) up -d).

nginx в depends_on ждёт balancer: service_healthy.

balancer в healthcheck ждёт :8079/ready (или сейчас :9098/ready, но HealthyBackends() == 0, пока gateway не готов).

Gateway не готов, потому что make deploy поднимает всё параллельно.

balancer unhealthy → nginx не стартует → make deploy падает (Ошибка 1).

Плюс balancer-2 — «dependency failed to start: container deployments-balancer-2 is unhealthy». Compose видит unhealthy и не запускает зависимые (nginx).

✅ Решение — два варианта
Вариант A (рекомендую): убрать condition: service_healthy у nginx
Файл: deployments/docker-compose.cluster.yml

Место: секция nginx, depends_on.

Сниппет:

yaml
  nginx:
    image: nginx:alpine
    container_name: deployments-nginx
    ports:
      - "80:80"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf:ro
    depends_on:
      - balancer
      - balancer-2
    networks:
      - event-horizon-net
    restart: unless-stopped
Почему: condition: service_healthy в compose — это блокирующий wait. На cold boot (когда gateway'и ещё стартуют) балансер всегда будет unhealthy какое-то время. nginx не должен блокироваться. Nginx стартанёт, а когда балансеры оживут — заработает. Плюс restart: unless-stopped у nginx — если он упадёт на старте (DNS не резолвится, потому что балансер не запущен), Docker перезапустит.

Минус: первые запросы к nginx до того, как балансер поднялся → 502. Но это секунды, и только на cold boot. Лучше, чем make deploy падает.

Плюс: healthcheck у балансеров оставляем — они будут healthy через ~10-15s, это полезно для docker ps и мониторинга.

Вариант B (жёстче): убрать healthcheck у балансеров, оставить depends_on простой
Файл: deployments/docker-compose.cluster.yml

Место: секции balancer, balancer-2.

Сниппет: убрать healthcheck: полностью, nginx → depends_on: - balancer - balancer-2.

Почему: healthcheck через wget нестабилен на cold boot. Мониторинг — через Prometheus (balancer_backend_healthy) + /ready на :9098. Docker healthcheck — не обязателен.

Минус: docker ps не покажет healthy/unhealthy. Но это косметика.

🎯 Мой выбор: Вариант A
Healthcheck балансера — оставляем (полезно).

Меняем test на :9098/ready.

Убираем condition: service_healthy у nginx → простой depends_on.

Почему не B: healthy в docker ps — полезный сигнал. А condition: service_healthy — это блокирующий wait, который на cold boot всегда ломает make deploy.


WE CHOSE "A"
