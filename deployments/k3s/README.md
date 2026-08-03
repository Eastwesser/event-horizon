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

### Запустить деплой в k3s

make deploy-k3s

### Проверить статус

kubectl get pods
kubectl get services
kubectl get ingress

### Посмотреть логи

# Все контейнеры в поде
kubectl logs deployment/event-horizon

# Конкретный контейнер
kubectl logs deployment/event-horizon -c auth
kubectl logs deployment/event-horizon -c billing
kubectl logs deployment/event-horizon -c game
kubectl logs deployment/event-horizon -c shop

### Подробная информация

kubectl describe pods

### Удалить деплой

make undeploy-k3s

---

## 🐛 ИЗВЕСТНЫЕ ПРОБЛЕМЫ

### 1. Поды падают с ошибкой NATS: no such host

**Проблема:** В k3s нет сервисов nats-1, nats-2, nats-3. NATS не развёрнут в Kubernetes.

**Решение:** Либо развернуть NATS в k3s, либо использовать внешний NATS.

**Текущий статус:** ⏳ Запланировано

### 2. Поды не видят PostgreSQL

**Проблема:** В k3s нет StatefulSet для PostgreSQL. БД развёрнута только в Docker Compose.

**Решение:** Добавить StatefulSet для PostgreSQL в k3s или использовать внешнюю БД.

**Текущий статус:** ⏳ Запланировано

### 3. Поды падают с CrashLoopBackOff

**Причина:** Приложение не может запуститься из-за отсутствия зависимостей (NATS, БД).

**Диагностика:**

kubectl logs deployment/event-horizon -c <имя_контейнера>

---

## 📊 СРАВНЕНИЕ С DOCKER COMPOSE

| Инфраструктура     | Статус                            | Команда           |
| :----------------- | :-------------------------------- | :---------------- |
| Docker Compose     | ✅ Работает                       | make deploy       |
| k3s (Kubernetes)   | 🟡 Запущен, часть подов падает    | make deploy-k3s   |

---

## 🧠 ДЛЯ СОБЕСЕДОВАНИЯ

«Я настроил k3s кластер и запустил деплой Event Horizon. 
Обнаружил, что для полноценной работы в Kubernetes нужно добавить StatefulSet для БД и настроить CoreDNS для резолвинга NATS. 
Это запланировано на следующий спринт.»

---

## 📝 ПОЛЕЗНЫЕ КОМАНДЫ

# Посмотреть все поды
kubectl get pods -o wide

# Посмотреть логи всех контейнеров в поде
kubectl logs deployment/event-horizon --all-containers

# Перезапустить под
kubectl rollout restart deployment/event-horizon

# Масштабировать (изменить количество реплик)
kubectl scale deployment/event-horizon --replicas=3

# Посмотреть события
kubectl get events --sort-by='.lastTimestamp'

# Войти в под (если есть shell)
kubectl exec -it deployment/event-horizon -c auth -- /bin/sh

---

## 🚧 TODO

- [ ] Добавить NATS в k3s
- [ ] Добавить StatefulSet для PostgreSQL
- [ ] Настроить CoreDNS
- [ ] Добавить Ingress для внешнего доступа
- [ ] Настроить автоматическое масштабирование (HPA)

Сделано с ❤️ для Event Horizon"