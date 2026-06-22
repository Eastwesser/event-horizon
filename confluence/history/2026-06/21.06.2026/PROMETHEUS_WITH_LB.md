Почему так:

Микросервисы в Docker → Prometheus видит их по именам контейнеров.
Gateway локально → Prometheus должен видеть его через host.docker.internal:9095.
Что дальше:

Проверь, что gateway локально отдает метрики:
bash
curl http://localhost:9095/metrics
Проверь, что Prometheus видит gateway:
Зайди в UI Prometheus: http://localhost:9090/targets

Проверь работу через balancer:
bash
curl http://localhost:8079/health
Почему balancer не работает? Потому что он ищет gateway-1, gateway-2, gateway-3 в Docker, а их нет (они не запущены в Docker). Но balancer должен проксировать на локальный gateway через 8081, 8082, 8083 — эти порты проброшены.

Проверь:

bash
curl http://localhost:8081/health
curl http://localhost:8082/health
curl http://localhost:8083/health
Если они работают — balancer должен работать. Если нет — balancer не может достучаться до gateway.