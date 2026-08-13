# CI/CD для backend-микросервисов (Event Horizon / red_mad_robot)

CI/CD — автоматическая проверка и доставка изменений. Senior объясняет стадии пайплайна, версионирование образов, миграции БД и работу с секретами. В EH: Taskfile/Makefile, Docker bin-образы, compose/k3s; **не пушить образы и не коммитить секреты без явной просьбы**.

## Цели пайплайна

- Быстрый feedback на PR (lint/test).
- Воспроизводимый артефакт (бинарь/image digest).
- Предсказуемый деплой с откатом.
- Безопасность: никаких секретов в git, минимальные права токенов CI.

## Типичные стадии (GitHub Actions / GitLab CI)

1. **Checkout** + кэш Go modules.
2. **Lint**: `golangci-lint` (в репо есть `.golangci.yml`).
3. **Unit tests**: `go test ./...` по модулям/`go.work`.
4. **Integration tests** (опционально): testcontainers / compose services.
5. **Build** бинарей сервисов (или multi-stage Docker build).
6. **Image build & push** (только с protected branch / manual approve).
7. **Migrate** job (с осторожностью) → **Deploy** (compose/k3s/helm).
8. **Smoke**: `/health` `/ready`, критичный gRPC/HTTP check.

Параллельте независимые job'ы; падайте early на lint/test.

## Образы как артефакты

- Тег: `gitsha` / semver; сохраняй digest для аудита.
- EH: `Dockerfile.*.bin` — CI сначала собирает бинарь, потом упаковывает тонкий образ.
- Не полагайся на `latest` в проде.
- Scan образа (Trivy/Grype) на критические CVE — gate перед prod.
- Push в registry только из CI с short-lived credentials (в EH локально — не пушить без запроса Emma).

## Миграции БД в CI/CD

- Миграции — часть релиза, не «руками на проде».
- Правила:
  - Backward compatible expand/contract (сначала add nullable/new table → выкатить код → удалить старое).
  - Избегать lock'ов на огромных таблицах без online-стратегии.
  - Отдельный migrate job с четкой версией; идемпотентность tool'а (goose/migrate/flyway-класс).
- EH: `pkg/migrator`, сервисы со своими schema; Inventory — чувствительный reference.
- Откат кода ≠ всегда откат миграции: готовь forward-fix.

## Секреты: жёсткие правила

- **Не коммитить** `.env`, приватные ключи, `JWT_SECRET` в манифестах с реальными значениями.
- Templates ок: `deployments/env/*.env.template`, примеры в README.
- CI secrets store (GitHub Secrets / GitLab CI variables) + K8s Secret / Vault на рантайме.
- Ротация: утечка в git history = revoke + rotate, не «удалить файл и забыть».
- Права: least privilege на deploy token; запрет force-push в main без политики.

## Деплой-стратегии

- Rolling (default k8s): постепенная замена Pod'ов.
- Blue/green или canary — для рискованных релизов gateway/billing.
- Feature flags — отделить выкатку кода от включения поведения.
- Compose на lab; k3s манифесты (`deployments/k3s/`) — упрощённый оркестраторный путь EH.

## Graceful shutdown и автоскейлинг (связь с CD)

- После выкладки старые Pod'ы получают SIGTERM: перестань принимать, добери in-flight, закрой пулы.
- HPA/автоскейл: по CPU или custom (RPS, lag) — но сначала правильные requests/limits.
- Не масштабируй blindly consumers без понимания порядка/партиций (Kafka) или durable semantics (NATS).

## Антипаттерны

- «Сборка на ноутбуке и docker push руками» как единственный процесс.
- Секреты в Dockerfile `ENV` или в git.
- Миграции mid-deploy без совместимости → downtime.
- Деплой без smoke и без возможности rollback image tag.
- Пропуск lint «чтобы быстрее влить».

## Чеклист релиза EH (практический)

- Тесты/lint зелёные.
- Образ(а) собраны с понятным тегом.
- Миграции применены/совместимы.
- `/ready` зелёный у затронутых сервисов.
- Метрики error rate/lag в норме первые N минут.
- Секреты только из secret store; `.env` не в коммите.

## Типичные вопросы на собесе

- Из каких стадий состоит ваш CI пайплайн и почему в таком порядке?
- Как версионируете Docker-образы и зачем digest?
- Как безопасно накатывать миграции в микросервисах?
- Где хранить секреты и что делать при утечке в git?
- Чем rolling отличается от canary?
- Как связан graceful shutdown с деплоем в k8s?
- Что автоматизировано в EH (Task/Make/Docker bin) и что нельзя делать без approve?
