# Docker для Event Horizon (senior / red_mad_robot)

Docker упаковывает сервис и зависимости в образ. На собесе ждут: multi-stage, минимальные runtime-образы, healthcheck, compose для локального кластера. В EH два стиля Dockerfile: полноценный multi-stage и **`Dockerfile.*.bin`** под уже собранный бинарь.

## Зачем контейнеры

- Одинаковый runtime на laptop / CI / k3s.
- Изоляция зависимостей (glibc, CA certs, timezone).
- Единица деплоя: image tag = артефакт релиза.
- Не замена правильной архитектуры: «нашинковать на контейнеры» ≠ микросервисы.

## Multi-stage build (классика)

- **Stage build**: `golang:1.xx`, `go mod download`, `go build -o /app/service`.
- **Stage runtime**: `gcr.io/distroless/static` / `alpine` / `scratch` + бинарь.
- Плюсы: маленький образ, нет компилятора и исходников в проде, быстрее pull.
- Типичные флаги: `CGO_ENABLED=0`, `-ldflags="-s -w"`, статическая линковка для scratch.
- Кэш слоёв: сначала `go.mod/go.sum`, потом исходники — быстрее CI.

## Dockerfile.*.bin в EH

- Идея: бинарь собирается снаружи (Task/Makefile/CI), Dockerfile только упаковывает файл.
- Плюсы: быстрый rebuild образа, единый pipeline компиляции, меньше логики в Docker.
- Минусы: образ «пустой» без предварительного `go build`; нужен дисциплинированный скрипт (`scripts/rebuild-phase1.sh`, Taskfile).
- Примеры: `Dockerfile.auth.bin`, `Dockerfile.shop.bin`, `Dockerfile.gateway.bin`, … для analytics/history/fulfillment/notification и др.
- На собесе: уметь объяснить trade-off multi-stage «всё в Docker» vs «bin + thin Dockerfile».

## Runtime hygiene

- Non-root user в финальном образе.
- Только нужные CA certificates и timezone data.
- Не копировать `.env` с секретами в образ; конфиг — env at runtime.
- Один процесс на контейнер (сервис слушает gRPC + metrics HTTP).
- Graceful shutdown: SIGTERM → stop accepting → drain → close PG/NATS.

## HEALTHCHECK и probes

- Dockerfile `HEALTHCHECK` полезен для compose; в k8s главные — liveness/readiness probes.
- EH паттерн: metrics HTTP с `GET /health` (liveness) и `GET /ready` (ping зависимостей).
- Healthcheck не должен быть тяжелее бизнес-нагрузки (не делай full migration check каждую секунду).

## Docker Compose (локальный кластер EH)

- `deployments/docker-compose.cluster.yml` — NATS 1..3, сервисы, зависимости.
- Per-domain compose под `deployments/compose/{auth,billing,shop,...}`.
- Env templates: `deployments/env/*.env.template` — **не коммитить реальные `.env`**.
- `depends_on` ≠ ready: нужен healthcondition / retry в приложении.
- Сети: внутренний bridge; наружу только gateway/порты мониторинга по необходимости.
- Volumes для PG/NATS/JetStream data — переживают рестарт compose.

## Образы и теги

- Тег = git sha или semver; `latest` только для демо.
- Registry (Docker Hub / GHCR): push по явному запросу (в EH — не пушить без просьбы Emma).
- Размер и CVE: периодический scan; обновляй base images.
- Multi-arch (amd64/arm64) — если команда на Apple Silicon + linux servers.

## Антипаттерны

- Компиляция в одном жирном `ubuntu` без multi-stage.
- Секреты в `ENV` слоях Dockerfile (остаются в history).
- `docker compose up` как единственный prod без оркестрации.
- Монтировать весь исходный код в прод-контейнер.
- Игнорировать init-порядок: приложение стартует до ready PG → бесконечный crash loop без backoff.

## Типичные вопросы на собесе

- Зачем multi-stage и чем distroless/scratch лучше ubuntu runtime?
- Как устроены `Dockerfile.*.bin` в EH и когда они выгодны?
- Чем `/health` отличается от `/ready` в контейнере?
- Как ускорить сборку через кэш слоёв Go-модулей?
- Почему `depends_on` недостаточно для старта сервиса?
- Куда класть секреты, если нельзя в образ и нельзя в git?
- Как сделать graceful shutdown Go-сервиса в Docker?
