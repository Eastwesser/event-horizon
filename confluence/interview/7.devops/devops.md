.

🐳 7. DEVOPS И ИНФРАСТРУКТУРА

❓ Как перейти с Docker Compose на k3s?
Ответ: Я бы делал это постепенно:

Сначала Helm-чарты — описываю каждый сервис в Helm (шаблонизированный Kubernetes манифест)

StatefulSets для БД — PostgreSQL и Redis должны сохранять данные

ConfigMap и Secrets — выношу конфиги из переменных окружения

Ingress + Service — настраиваю маршрутизацию

GitOps (ArgoCD) — автоматический деплой из репозитория

В Event Horizon: Docker Compose работает, но я уже закладываю конфиги, которые легко перенести на k3s.

Dockerfile для Shop:

dockerfile
FROM scratch
COPY services/shop/shop-service /shop-service
EXPOSE 50055
CMD ["/shop-service"]



❓ Что такое Helm-чарты?
Ответ: Helm — это package manager для Kubernetes. Чарт — это набор шаблонов, которые описывают deployment, service, configmap, secrets.

Зачем:

Один чарт для всех окружений (dev/staging/prod)

Легко менять версии

Возможность отката (helm rollback)

Пример структуры:

text
shop-chart/
├── Chart.yaml
├── values.yaml
├── templates/
    ├── deployment.yaml
    ├── service.yaml
    └── configmap.yaml



❓ Как бы реализовал canary deployment?
Ответ: Canary — это постепенный rollout новой версии.

В k3s это делается через:

Istio — управление трафиком

Argo Rollouts — продвинутые стратегии

Схема:

Запускаю новую версию с 10% трафика

Мониторю ошибки и latency

Если всё ок — увеличиваю до 50%

Через час — 100%

В Prometheus: Алерт, если error rate > 1% на canary.



❓ Как собираешь multi-stage Docker-образы?
Ответ: Multi-stage позволяет разделить сборку и финальный образ.

Dockerfile для Auth:

dockerfile
# Stage 1: Сборка
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o auth-service ./cmd/main.go

# Stage 2: Финальный образ
FROM alpine:latest
COPY --from=builder /app/auth-service /auth-service
EXPOSE 50051
CMD ["/auth-service"]
Размер:

Без multi-stage: ~500 MB (Go + зависимости)

С multi-stage: ~15 MB (только бинарник)

В Event Horizon: Я использую FROM scratch для Shop, это даёт образ ~20 MB.



❓ Как настроить HPA для генерации картинок?
Ответ: HPA (Horizontal Pod Autoscaler) масштабирует поды по метрикам.

Для очереди NATS:

Использую custom metrics (Prometheus adapter)

Смотрю длину очереди: nats_consumer_lag{consumer="generator"}

yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
spec:
  metrics:
  - type: Pods
    pods:
      metric:
        name: nats_consumer_lag
      target:
        type: AverageValue
        averageValue: 100


❓ Как настроить CI/CD в GitHub Actions?
Ответ:

yaml
name: Deploy to k3s
on:
  push:
    branches: [main]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - name: Build and push Docker images
      run: |
        for service in auth billing game shop gateway balancer; do
          docker build -t eastwesser/$service:latest -f Dockerfile.$service .
          docker push eastwesser/$service:latest
        done
    
    - name: Deploy to k3s
      run: |
        kubectl set image deployment/auth auth=eastwesser/auth:latest
        kubectl rollout status deployment/auth



❓ Что если прод упал во время деплоя?
Ответ: План действий:

Откат — kubectl rollout undo deployment/auth

Проверить логи — kubectl logs deployment/auth --tail=100

Проверить события — kubectl describe pod

Важно: В Helm можно настроить rollback на предыдущую версию.


❓ Как передаешь секреты в Kubernetes?
Ответ:

Secrets — kubectl create secret generic db-password --from-literal=password=...

Vault — для более сложных сценариев (внешнее хранилище)

В k3s:

yaml
apiVersion: v1
kind: Secret
metadata:
  name: db-secret
type: Opaque
data:
  password: cGFzc3dvcmQ= # base64


❓ Service — ClusterIP vs LoadBalancer?
Ответ:

ClusterIP — доступ только внутри кластера (для внутренних сервисов)

LoadBalancer — доступ снаружи (например, для Gateway)

В Event Horizon: Gateway будет LoadBalancer, всё остальное — ClusterIP.



❓ Зачем Ingress, если есть LoadBalancer?
Ответ: Ingress — это маршрутизация на уровне HTTP.

Что делает Ingress:

Маршрутизация по пути (/api/auth → Auth, /api/game → Game)

SSL-терминация

Rate Limiting

Если только LoadBalancer:

Каждый сервис получает свой внешний IP

Нет маршрутизации

В Event Horizon: Ingress будет маршрутизировать все запросы к Gateway.



🎯 КЛЮЧЕВЫЕ ТЕЗИСЫ ДЛЯ СОБЕСА
Переход на k3s — начинаем с Helm-чартов и ConfigMap

Multi-stage — собираем бинарник в одной стадии, копируем в чистый образ

HPA — масштабируем по длине очереди NATS

CI/CD — GitHub Actions → Docker Hub → k3s

Роллбек — kubectl rollout undo

Ingress — маршрутизация HTTP-запросов

