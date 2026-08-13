# Kubernetes / k3s для Event Horizon

Kubernetes оркестрирует контейнеры: желаемое состояние, самолечение, сервис-дискавери, конфиги/секреты. На собесе red_mad_robot ждут базовый словарь объектов и probes. В EH для упрощённого деплоя есть манифесты **k3s** (`deployments/k3s/`).

## Основные объекты

- **Pod**: минимальная единица (один или несколько контейнеров, shared network/volume). Эфемерен.
- **Deployment**: декларативные реплики Pod'ов, rolling update, rollback. Контроллер поддерживает ReplicaSet.
- **Service**: стабильный VIP/DNS для набора Pod'ов (ClusterIP / NodePort / LoadBalancer).
- **Ingress**: L7-вход (HTTP host/path) → Service; TLS termination часто здесь.
- **ConfigMap**: несекретный конфиг; **Secret**: чувствительные данные (base64, не шифрование само по себе — нужен encryption at rest / external KMS).
- **HPA**: автоскейл реплик по CPU/custom metrics.
- **Job/CronJob**: миграции и разовые задачи (осторожно с lock'ами).

## Deployment: практика

- `replicas`, `resources.requests/limits` — без requests планировщик «угадывает» плохо.
- Rolling update: `maxUnavailable` / `maxSurge`; readiness мешает слать трафик неготовым Pod'ам.
- Labels + selectors должны совпадать (частая ошибка манифеста).
- Image tag immutable (sha), не только `latest`.
- В EH примере `deployments/k3s/deployment.yml` — несколько контейнеров/сервисов (auth, billing, …) с env и `secretKeyRef` для JWT; для учёбы ок, для prod лучше 1 сервис = 1 Deployment.

## Service и Ingress

- ClusterIP — внутри кластера (gateway → auth:50051).
- Ingress / API Gateway снаружи: TLS, маршрутизация на gateway service.
- Не светить все gRPC порты наружу без нужды.
- DNS имя: `http://service.namespace.svc.cluster.local` (коротко `service` в той же ns).

## Probes: liveness, readiness, startup

- **Liveness**: процесс мёртв/завис → kubelet рестартит контейнер. Только дешёвый `/health`.
- **Readiness**: не готов принимать трафик (нет PG/NATS) → убрать endpoints из Service. `/ready` с ping зависимостей.
- **StartupProbe**: для медленного старта, чтобы liveness не убил раньше времени.
- Антипаттерн: readiness = тяжёлый бизнес-запрос; liveness зависит от даунстрима (каскадные рестарты).

## Config и секреты в EH k3s

- `deployments/k3s/secret.yml` — JWT и пр.; не коммитить реальные прод-значения.
- Env из Secret через `valueFrom.secretKeyRef` (как JWT_SECRET у auth в deployment.yml).
- NATS_URL: список `nats://nats-1:4222,...`; DB_* и REDIS_ADDR — отдельные сервисы/хосты.
- Смена Secret → нужен rollout Pod'ов (env не hot-reload сам).

## k3s note для Event Horizon

- **k3s** — лёгкий Kubernetes (один бинарь, удобен для lab/edge/homelab).
- Манифесты EH рассчитаны на упрощённый сценарий, не на полный prod multi-cluster.
- Локальная разработка чаще на Docker Compose cluster; k3s — шаг к оркестрации.
- На собесе: «используем k3s для упрощённого деплоя EH; те же объекты Deployment/Service/Secret, что в полном k8s».

## Релизы и устойчивость

- Canary / blue-green: постепенно лить трафик на новый ReplicaSet (Ingress weights / mesh).
- PDB (PodDisruptionBudget) — не уронить все реплики при drain.
- Graceful shutdown: `terminationGracePeriodSeconds` + обработка SIGTERM в Go.
- Миграции БД: отдельный Job/init до/во время выкладки с backward-compatible schema.

## Типичные вопросы на собесе

- Чем Pod отличается от Deployment и зачем Service?
- Разница liveness vs readiness vs startup probe?
- Как передать секреты в Pod и почему Secret ≠ шифрование?
- Что такое rolling update и как откатиться?
- Зачем requests/limits?
- Чем k3s отличается от «большого» Kubernetes и как это используется в EH?
- Как безопасно выкатить breaking change API в кластере?
