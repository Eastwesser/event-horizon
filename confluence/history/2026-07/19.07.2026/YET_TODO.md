❌ Что из TODO НЕ РЕАЛИЗОВАНО (надо делать)

🔥 Критично (для стабильности)
Индексы в БД — добавить во все таблицы (highscores, user_currencies, inventory, purchases)

Rate Limiter в Gateway (раскомментировать и настроить)

Добить K6 тесты — p95 < 500ms, ошибки < 1%

🟡 Важно (для масштабирования)
Юнит-тесты (покрытие 70%+)

Graceful shutdown во всех сервисах

OpenAPI / Swagger для всех эндпоинтов

README.md для каждого сервиса

Антифрод (ограничение 1 фенечка/месяц, fingerprint, капча)

🟢 Новые сервисы
Payment (Boosty/Stripe) — реальные платежи

Notification (Telegram/Email) — уведомления

Analytics (ClickHouse) — DAU, MAU, Retention

⚙️ DevOps
k3s (Kubernetes) — Helm-чарты, Ingress

CI/CD — GitHub Actions

Ansible — автоматизация деплоя

Service Discovery — Consul

📂 Что из техдолга осталось
Из папки tech_debt/:

HARDCODED_PORTS.md — порты хардкодом в коде (надо вынести в config)

MIGRATIONS_AUTO.md — автоматизация миграций (сейчас руками)

PERFORMANCE_TODO.md — оптимизация запросов, индексы

DEVOPS_LIST.md — CI/CD, мониторинг, алерты

DEBT_LIST.md — общий список техдолга

🎯 Что предлагаю делать сейчас (по приоритету)
Индексы в БД — 30 минут, даст сразу видимый эффект в K6

Rate Limiter — 15 минут (раскомментировать код)

K6 тесты — добить до p95 < 500ms (уже почти готово)

README для каждого сервиса — можно сделать шаблон и заполнить за час

Антифрод — начать с ограничения на 1 фенечку/месяц