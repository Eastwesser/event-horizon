🔮 Планы на следующие спринты

🔥 Ближайшие задачи (1–2 недели)

    Нагрузочное тестирование (k6) — прогнать все сценарии, замерить RPS, latency

    Рефакторинг кода — убрать хардкод, добавить структурные логи, комментарии

    Rate Limiter — раскомментировать, настроить лимиты (10/сек на пользователя)

    Документация API — OpenAPI/Swagger, README для каждого сервиса

    Тесты — юнит-тесты (≥70% покрытия), интеграционные тесты (testcontainers)

⚙️ DevOps (1–2 недели)

    CI/CD — GitHub Actions: сборка, пуш в Docker Hub, деплой через SSH

    Ansible — автоматизация установки Docker, копирования бинарников

    k3s (Kubernetes) — Helm-чарты, Ingress (Traefik), горизонтальное масштабирование

Service Discovery — Consul для регистрации сервисов

🧩 Новые сервисы (1–2 месяца)

Сервис	Назначение	Порт (gRPC)

Shop	        Магазин за билетики	                5055
Notification	Push, Email, Telegram	            5056
Analytics	    DAU, MAU, Retention (ClickHouse)	5057
Payment	        Реальные платежи (Boosty/Stripe)	5058

🎮 Игровой контент

    Добавить игры: flappy, towers, memory

    Уровни сложности (1–20)

    Достижения (achievements)

    Блинопекарня (магазин за лампочки)

🧠 Устойчивость

    Circuit Breaker + Bulkhead

    Retry с джиттером

    Graceful shutdown

    Алерты в Telegram (Alertmanager)

🧑‍💻 Команда
Backend & DevOps: Денис Матвеев (Eastwesser)

Архитектура: Микросервисная, событийно-ориентированная

Деплой: Docker Compose (сейчас) → k3s (в планах)

📦 Версия
Текущая: v1.0.3 (05.07.2026)
Следующий релиз: v1.1.0 (план — 10.07.2026)

⭐ Если проект полезен
⭐ Поставь звезду на GitHub
🐛 Создай Issue
📬 Напиши мне: eastwesser@gmail.com

Event Horizon — играй, соревнуйся, побеждай! 🚀