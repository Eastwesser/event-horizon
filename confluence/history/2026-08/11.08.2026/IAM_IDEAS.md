📊 Текущая ситуация с безопасностью:
Компонент	Доступ	Кто может смотреть
Prometheus (9090)	Открыт без аутентификации	Любой, кто знает IP/порт
Grafana (3000)	admin/admin (по умолчанию)	Любой, кто знает логин/пароль
Jaeger (16686)	Открыт без аутентификации	Любой, кто знает IP/порт
Gateway API	JWT токен	Только авторизованные пользователи
Базы данных	Закрыты внутри Docker сети	Только сервисы
🚨 Риски:
Prometheus — показывает метрики (RPS, ошибки, задержки). Это не критично, но может дать злоумышленнику информацию о нагрузке и уязвимостях.

Jaeger — показывает трассировки запросов. Может раскрыть внутреннюю структуру API и данные пользователей.

Grafana — если кто-то зайдёт под admin/admin, сможет смотреть дашборды и настраивать источники данных.

🔧 Решения:
1. Prometheus — добавить basic auth или прокси с аутентификацией:
yaml
# В docker-compose.cluster.yml для prometheus:
prometheus:
  image: prom/prometheus:latest
  command:
    - '--config.file=/etc/prometheus/prometheus.yml'
    - '--web.enable-admin-api'
    - '--web.config.file=/etc/prometheus/web.yml'  # ← добавить
  volumes:
    - ./prometheus/prometheus.yml:/etc/prometheus/prometheus.yml:ro
    - ./prometheus/web.yml:/etc/prometheus/web.yml:ro  # ← добавить
Файл web.yml:

yaml
basic_auth_users:
  admin: $2y$10$... # хеш пароля (можно сгенерировать через htpasswd)
2. Grafana — сменить пароль admin:
bash
# Войти в Grafana и сменить пароль admin
# Или через переменные окружения:
- GF_SECURITY_ADMIN_USER=admin
- GF_SECURITY_ADMIN_PASSWORD=<новый_пароль>
3. Jaeger — добавить nginx-прокси с basic auth
4. Gateway — добавить роли пользователей (admin/user/author)
🛡️ План по ролям для будущего:
Роль	Что может делать
User	Играть, покупать, смотреть свой профиль
Author	Всё что User + создавать/редактировать товары
Admin	Всё что Author + управлять пользователями, модерировать, смотреть метрики
📋 Что можно сделать уже сейчас (быстро):
bash
# 1. Сменить пароль Grafana
docker exec -it event-horizon-grafana grafana-cli admin reset-admin-password <новый_пароль>

# 2. Перезапустить Grafana
docker restart event-horizon-grafana

# 3. Закрыть порты 9090, 16686 от внешнего доступа (если не нужно)
# В docker-compose убрать проброс портов для prometheus и jaeger
# Или добавить nginx-прокси с basic auth