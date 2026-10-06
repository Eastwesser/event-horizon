# ☸️ K3S — Event Horizon

## 📌 Что это
Kubernetes кластер (k3s) для Event Horizon. Развёрнут параллельно с Docker Compose.

---

## 🔧 УСТАНОВКА И НАСТРОЙКА

### 1. Установка k3s

curl -sfL https://get.k3s.io | sh -

### 2. Проверка статуса

sudo systemctl status k3s

### 3. Настройка kubectl

# Создать папку для конфига
mkdir -p ~/.kube

# Скопировать конфиг
sudo cp /etc/rancher/k3s/k3s.yaml ~/.kube/config

# Дать права пользователю
sudo chown denismatveev:denismatveev ~/.kube/config
chmod 600 ~/.kube/config

# Указать путь к конфигу
export KUBECONFIG=~/.kube/config
echo 'export KUBECONFIG=~/.kube/config' >> ~/.bashrc
source ~/.bashrc

### 4. Проверка

kubectl get nodes
kubectl get pods -A

---

## 🚀 ДЕПЛОЙ

### Helm (preferred — Wave 4.5)

```bash
make deploy-k3s          # helm upgrade --install eh ./deployments/helm/event-horizon
make undeploy-k3s
make helm-template-k3s   # dry-run render
```

Chart: [`../helm/event-horizon/`](../helm/event-horizon/). Raw YAML below remains a reference snapshot.

### Legacy kubectl apply

make deploy-k3s

### Проверить статус

kubectl get pods
kubectl get services
kubectl get ingress

### Посмотреть логи

# Все контейнеры в поде
kubectl logs deployment/eh-event-horizon

# Конкретный контейнер
kubectl logs deployment/eh-event-horizon -c auth
kubectl logs deployment/eh-event-horizon -c billing
kubectl logs deployment/eh-event-horizon -c game
kubectl logs deployment/eh-event-horizon -c shop

### Подробная информация

kubectl describe pods

### Удалить деплой

make undeploy-k3s

---

## 🐛 ИЗВЕСТНЫЕ ПРОБЛЕМЫ

### 1. NATS / Postgres in-cluster (Track C)

**Было:** поды падали с `no such host` — в k3s не было `nats-1..3` / `postgres*`.

**Сейчас:** Helm chart умеет data plane:

```bash
make deploy-k3s-dataplane
# или
helm upgrade --install eh ./deployments/helm/event-horizon --set dataPlane.enabled=true
```

Рендерит StatefulSets + headless Services с DNS-именами как в `values.yaml` (`natsURL`, `db.hosts`). Redis по-прежнему внешний / compose.

### 2. Поды падают с CrashLoopBackOff

**Причина:** без data plane (или без внешнего NATS/PG/Redis) зависимости недоступны.

**Диагностика:**

kubectl logs deployment/eh-event-horizon -c <имя_контейнера>

---

## 📊 СРАВНЕНИЕ С DOCKER COMPOSE

| Инфраструктура     | Статус                            | Команда                    |
| :----------------- | :-------------------------------- | :------------------------- |
| Docker Compose     | ✅ Day-to-day                     | make deploy                |
| k3s app only       | 🟡 нужен внешний data plane       | make deploy-k3s            |
| k3s + data plane   | ✅ NATS + Postgres STS in chart   | make deploy-k3s-dataplane  |

---

## 🧠 ДЛЯ СОБЕСЕДОВАНИЯ

«Я вынес app в Helm (Wave 4.5), а data plane (NATS cluster + per-service Postgres) добавил как opt-in StatefulSets в том же chart — DNS-имена совпадают с compose/`natsURL`, чтобы не переписывать сервисы.»

---

## 📝 ПОЛЕЗНЫЕ КОМАНДЫ

# Посмотреть все поды
kubectl get pods -o wide

# Data plane
kubectl get sts,svc | grep -E 'nats-|postgres'

# Посмотреть логи всех контейнеров в поде
kubectl logs deployment/eh-event-horizon --all-containers

# Перезапустить под
kubectl rollout restart deployment/eh-event-horizon

# Масштабировать (изменить количество реплик)
kubectl scale deployment/eh-event-horizon --replicas=3

# Посмотреть события
kubectl get events --sort-by='.lastTimestamp'

# Войти в под (если есть shell)
kubectl exec -it deployment/eh-event-horizon -c auth -- /bin/sh

---

## 🚧 TODO

- [x] Helm chart wrapping app manifests (Wave 4.5 — `deployments/helm/event-horizon`)
- [x] NATS StatefulSets in Helm (`dataPlane.enabled`)
- [x] Postgres StatefulSets in Helm (`dataPlane.enabled`, 7 DBs from `db.hosts`)
- [x] Ingress для внешнего доступа (Traefik / chart)
- [ ] Redis in-cluster (optional; not Track C checkbox)
- [ ] Настроить автоматическое масштабирование (HPA)
- [ ] Patroni HA (separate — `deployments/patroni/`)

Сделано с ❤️ для Event Horizon
