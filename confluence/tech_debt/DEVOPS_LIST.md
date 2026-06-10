## DevOps CheckList — EventHorizon

### ✅ Уже есть
- [x] Docker Compose для всех сервисов
- [x] Graceful shutdown
- [x] Volumes для БД (постоянное хранение)
- [x] Healthchecks в docker-compose
- [x] Логи в файлы (/tmp/*.log)

### ❌ Нужно добавить / проверить

#### Инфраструктура
- [ ] **Бэкапирование БД** — PostgreSQL (pg_dump, WAL归档)
- [ ] **Автозапуск DNS** — после рестарта сервера
- [ ] **Мониторинг и алерты** — Prometheus + Grafana + Alertmanager
- [ ] **CI/CD** — GitHub Actions / GitLab CI
- [ ] **Ручной деплой** — документирован ли процесс

#### Сеть и безопасность
- [ ] **БД на разных серверах** — сейчас всё на одной VM
- [ ] **Приватная сеть / VPC** — изолированы ли сервисы
- [ ] **БД недоступна извне** — закрыты порты (5460-5463)
- [ ] **Reverse Proxy (nginx/Envoy)** — перед Gateway

#### Миграции
- [ ] **Dual Writes** — стратегия миграции без даунтайма
- [ ] **Миграции для всех сервисов** — Auth, Game, Billing, Leaderboard
- [ ] **Откат миграций** — документирован ли процесс

#### High Availability
- [ ] **Балансировщики нагрузки** — LB перед Gateway
- [ ] **Мастер-реплики** — для PostgreSQL (чтение с реплик)
- [ ] **Uptime 99%** — 87.6 часов в год downtime максимум
- [ ] **NATS кластер** — 3 ноды (в планах)

#### Хранилище образов
- [ ] **Docker Hub** — бизнес-аккаунт / Harbor / Nexus
- [ ] **Версионирование образов** — тегирование

#### Логи и Observability
- [ ] **Grafana Loki** — или ELK для логов
- [ ] **Постмортемы** — документированы в Confluence
- [ ] **Трейсинг** — Jaeger (опционально)

#### CI/CD Pipeline (полный цикл)
- [ ] Линтеры (golangci-lint, ESLint)
- [ ] Проверки безопасности (govulncheck, npm audit)
- [ ] Сборка (go build, npm run build)
- [ ] Пуш в реестр (Docker Hub)
- [ ] Деплой на dev
- [ ] Прогон тестов (unit, integration, e2e)
- [ ] Деплой на stage
- [ ] Ручная проверка
- [ ] Деплой в production

#### Оркестрация
- [ ] **Kubernetes** — в планах (k3s / kind)
- [ ] **k6** — нагрузочное тестирование (уже есть)

---

## Бэкенд — Кастомный ник в лидерборде

**Статус:** ❌ Пока не реализовано

### Что нужно сделать:

| Компонент | Действие | Файл |
|-----------|----------|------|
| Proto (game) | Добавить поле `nickname` | `services/game/proto/game.proto` |
| Game service | Пробросить nickname в NATS | `services/game/internal/service/game_service.go` |
| Proto (leaderboard) | Добавить поле `nickname` | `services/leaderboard/proto/leaderboard.proto` |
| Leaderboard repo | Сохранять nickname в Redis | `services/leaderboard/internal/repository/redis_repo.go` |
| Leaderboard service | Возвращать nickname в API | `services/leaderboard/internal/service/leaderboard_service.go` |
| Gateway | Пробросить nickname из JSON в gRPC | `services/gateway/cmd/main.go` |

### Оценка: 🟡 Средняя (~2-3 часа)