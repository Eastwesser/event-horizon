3. 502 Nginx — корневая причина: Nginx ретраит 502, но не 503/504
nginx.conf:

nginx
proxy_next_upstream               error timeout http_502;
proxy_next_upstream_tries         2;
proxy_next_upstream_timeout       3s;
proxy_connect_timeout 2s;
proxy_send_timeout    30s;
proxy_read_timeout    30s;
proxy_read_timeout 30s — вот откуда 36s p99. Nginx ждёт 30 секунд ответа от Balancer, потом 502. А Balancer ждёт Gateway, Gateway ждёт gRPC-сервис. Под P=200 вся цепочка висит 30 секунд, потом 502.

Плюс: proxy_next_upstream ... http_502 — Nginx ретраит 502 на второй Balancer. Это дублирует запрос, если первый Balancer уже начал обработку. Для read path — ок, для purchase — двойное списание? Нет, purchase идемпотентен (saga), но всё равно лишняя нагрузка.

Фикс:

Снизить proxy_read_timeout до 5–10s — быстрый fail вместо 30s зависания.

Убрать http_502 из proxy_next_upstream — ретраить только error timeout (соединение не установлено), не 502 (сервер ответил ошибкой).

proxy_connect_timeout 1s — быстрее отсекать мёртвые Balancer'ы.

Приоритет: P1. 36s p99 — это не «медленно», это «зависло на 30 секунд». Пользователь ушёл.