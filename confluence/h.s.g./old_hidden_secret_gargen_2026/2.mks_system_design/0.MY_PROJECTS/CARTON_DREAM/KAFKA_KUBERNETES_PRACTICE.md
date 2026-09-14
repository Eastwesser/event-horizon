# 🚀 Carton Dream — Практика Kafka + Kubernetes

## 🎯 Что ты получишь

Этот проект — **идеальная песочница** для практики:
- ✅ **Kafka** — producer, consumer, topics, consumer groups
- ✅ **Kubernetes** — StatefulSet, Services, ConfigMaps, deployments
- ✅ **gRPC** — межсервисная коммуникация
- ✅ **RabbitMQ** — параллельно с Kafka для сравнения
- ✅ **PostgreSQL** — 5 баз данных, миграции
- ✅ **Мониторинг** — Prometheus + Grafana

---

## 📊 Архитектура Kafka в проекте

```
┌─────────────────────────────────────────────────────────┐
│                   Admin Service                          │
│              (Kafka Producer & Consumer)                 │
└───────────────────┬─────────────────────────────────────┘
                    │
                    ▼
        ┌───────────────────────┐
        │   Kafka Cluster       │
        │   (3 brokers in K8s)  │
        │   - Port: 9092/9093   │
        └───────────┬───────────┘
                    │
        ┌───────────┴───────────┐
        │                       │
    ┌───▼────┐            ┌─────▼─────┐
    │ Orders │            │  Audit    │
    │ Topic  │            │  Topic    │
    └────────┘            └───────────┘
        │                       │
        ▼                       ▼
  ┌──────────┐          ┌──────────┐
  │ Orders   │          │ Analytics│
  │ Service  │          │ Service  │
  └──────────┘          └──────────┘
```

---

## 🔧 Запуск проекта (Docker Compose)

### Шаг 1: Подготовка

```bash
cd /home/denismatveev/Downloads/musashi/go-go-go-2025/petprojects/carton-dream

# Проверь наличие .env файла
ls -la .env

# Если нет — создай из примера
cp .env.example .env

# Заполни минимум:
nano .env
```

**Минимальные переменные**:
```bash
# JWT
JWT_SECRET=your_super_secret_key_min_32_characters_long

# Telegram Bot (опционально, если хочешь тестировать бота)
TELEGRAM_API_KEY=your_telegram_bot_token

# DeepSeek AI (опционально)
DEEPSEEK_API_KEY=sk-your_key

# Админка (для Kafka)
ADMIN_SECRET=change_this_admin_secret
```

### Шаг 2: Запуск всего стека

```bash
# Запуск ВСЕХ сервисов (включая Kafka, Zookeeper, RabbitMQ)
docker-compose up -d

# Следи за логами
docker-compose logs -f

# Проверь статус
docker-compose ps
```

**Что запустится**:
- ✅ Zookeeper (порт 2181)
- ✅ Kafka (порт 9092/9093)
- ✅ Kafka UI (порт 8080)
- ✅ RabbitMQ (порт 5672, UI: 15672)
- ✅ PostgreSQL (5 баз: 5432-5436)
- ✅ 5 микросервисов (Auth, Profile, Orders, Payment, Notification)
- ✅ Admin Service (с Kafka интеграцией)
- ✅ Telegram Bot
- ✅ Frontend (Angular)

### Шаг 3: Проверка Kafka

```bash
# Kafka UI (лучший способ)
open http://localhost:8080

# Проверь Kafka внутри контейнера
docker exec -it carton-kafka kafka-topics --list --bootstrap-server localhost:9092

# Создай тестовый топик
docker exec -it carton-kafka kafka-topics --create \
  --topic test-topic \
  --bootstrap-server localhost:9092 \
  --partitions 3 \
  --replication-factor 1

# Отправь сообщение (producer)
docker exec -it carton-kafka kafka-console-producer \
  --broker-list localhost:9092 \
  --topic test-topic

# Получи сообщения (consumer)
docker exec -it carton-kafka kafka-console-consumer \
  --bootstrap-server localhost:9092 \
  --topic test-topic \
  --from-beginning
```

---

## 🎓 Практика #1: Kafka Producer & Consumer

### Задача 1: Отправь событие через Admin Service

```bash
# Admin Service слушает на порту 8082
# Создай тестовое событие:

curl -X POST http://localhost:8082/api/v1/kafka/publish \
  -H "Content-Type: application/json" \
  -d '{
    "topic": "orders.created",
    "key": "order-123",
    "value": {
      "order_id": "123",
      "user_id": "user-456",
      "status": "pending",
      "created_at": "2025-02-16T12:00:00Z"
    }
  }'
```

### Задача 2: Реализуй Consumer для нового топика

Создай файл `backend/orders/internal/kafka/test_consumer.go`:

```go
package kafka

import (
	"context"
	"log"
	
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type TestConsumer struct {
	consumer *kafka.Consumer
}

func NewTestConsumer(brokers string) (*TestConsumer, error) {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": brokers,
		"group.id":          "test-consumer-group",
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		return nil, err
	}
	
	return &TestConsumer{consumer: c}, nil
}

func (tc *TestConsumer) Start(ctx context.Context, topic string) error {
	err := tc.consumer.Subscribe(topic, nil)
	if err != nil {
		return err
	}
	
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			msg, err := tc.consumer.ReadMessage(-1)
			if err != nil {
				log.Printf("Consumer error: %v\n", err)
				continue
			}
			
			log.Printf("Received message: %s = %s\n", 
				string(msg.Key), string(msg.Value))
		}
	}
}

func (tc *TestConsumer) Close() {
	tc.consumer.Close()
}
```

**Задание**: Реализуй логику обработки сообщений с десериализацией JSON.

---

## ☸️ Практика #2: Kubernetes Deployment

### Шаг 1: Установи Minikube/Kind (если нет)

```bash
# Minikube
curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-linux-amd64
sudo install minikube-linux-amd64 /usr/local/bin/minikube

# Запуск
minikube start --cpus=4 --memory=8192

# Или используй Kind (легковеснее)
kind create cluster --name carton-dream
```

### Шаг 2: Деплой Kafka в Kubernetes

```bash
# Создай namespace
kubectl create namespace carton-dream

# Деплой Zookeeper + Kafka
kubectl apply -f k8s/infrastructure/kafka.yaml

# Проверь статус
kubectl get pods -n carton-dream -w

# Посмотри логи Kafka
kubectl logs -n carton-dream kafka-0 -f
```

**Что происходит**:
- StatefulSet создает 3 Kafka брокера (`kafka-0`, `kafka-1`, `kafka-2`)
- Каждый брокер имеет свой PersistentVolume (10 Gi)
- Service `kafka` создает headless service для StatefulSet

### Шаг 3: Деплой всех сервисов

```bash
# ConfigMap с переменными
kubectl apply -f k8s/base/configmap.yaml

# Databases
kubectl apply -f k8s/databases/

# Services
kubectl apply -f k8s/services/

# Ingress (если есть)
kubectl apply -f k8s/infrastructure/ingress.yaml
```

### Шаг 4: Подключись к Kafka внутри K8s

```bash
# Port-forward для доступа снаружи
kubectl port-forward -n carton-dream svc/kafka 9092:9092

# Теперь можешь подключиться с localhost:9092
kafka-console-consumer --bootstrap-server localhost:9092 --topic orders.created
```

---

## 🎓 Практика #3: Helm Charts

### Шаг 1: Установи Helm

```bash
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
```

### Шаг 2: Используй готовый Helm Chart

```bash
cd helm/carton-dream

# Посмотри values.yaml
cat values.yaml

# Установи релиз
helm install carton-dream . \
  --namespace carton-dream \
  --create-namespace \
  --set kafka.enabled=true \
  --set kafka.replicas=3

# Проверь статус
helm status carton-dream -n carton-dream

# Обнови конфигурацию
helm upgrade carton-dream . \
  --namespace carton-dream \
  --set kafka.replicas=5
```

### Шаг 3: Настрой Values для практики

Отредактируй `helm/carton-dream/values.yaml`:

```yaml
kafka:
  enabled: true
  replicas: 3
  resources:
    requests:
      memory: "2Gi"
      cpu: "1"
    limits:
      memory: "4Gi"
      cpu: "2"
  
  persistence:
    enabled: true
    size: 20Gi
  
  topics:
    - name: orders.created
      partitions: 6
      replication: 3
    - name: audit.events
      partitions: 12
      replication: 3
```

---

## 🧪 Практика #4: Kafka Topics & Partitions

### Задача: Создай топики для разных сценариев

```bash
# Топик для заказов (много сообщений, нужна скорость)
docker exec -it carton-kafka kafka-topics --create \
  --topic orders.created \
  --bootstrap-server localhost:9092 \
  --partitions 6 \
  --replication-factor 1

# Топик для аудита (все события, долгое хранение)
docker exec -it carton-kafka kafka-topics --create \
  --topic audit.events \
  --bootstrap-server localhost:9092 \
  --partitions 12 \
  --replication-factor 1 \
  --config retention.ms=604800000  # 7 дней

# Топик для аналитики (batch processing)
docker exec -it carton-kafka kafka-topics --create \
  --topic analytics.metrics \
  --bootstrap-server localhost:9092 \
  --partitions 3 \
  --replication-factor 1 \
  --config cleanup.policy=compact
```

### Проверка партиций

```bash
# Посмотри детали топика
docker exec -it carton-kafka kafka-topics --describe \
  --topic orders.created \
  --bootstrap-server localhost:9092

# Вывод:
# Topic: orders.created   PartitionCount: 6       ReplicationFactor: 1
# Partition: 0    Leader: 1       Replicas: 1     Isr: 1
# ...
```

---

## 🎓 Практика #5: Consumer Groups

### Задача: Создай 3 consumer в одной группе

Создай 3 терминала и запусти:

**Terminal 1**:
```bash
docker exec -it carton-kafka kafka-console-consumer \
  --bootstrap-server localhost:9092 \
  --topic orders.created \
  --group order-processors \
  --from-beginning
```

**Terminal 2**:
```bash
docker exec -it carton-kafka kafka-console-consumer \
  --bootstrap-server localhost:9092 \
  --topic orders.created \
  --group order-processors \
  --from-beginning
```

**Terminal 3**:
```bash
docker exec -it carton-kafka kafka-console-consumer \
  --bootstrap-server localhost:9092 \
  --topic orders.created \
  --group order-processors \
  --from-beginning
```

**Terminal 4 (producer)**:
```bash
docker exec -it carton-kafka kafka-console-producer \
  --broker-list localhost:9092 \
  --topic orders.created

# Отправь 10 сообщений:
message 1
message 2
...
message 10
```

**Наблюдение**: Сообщения распределятся между 3 consumers (load balancing).

### Проверь consumer group

```bash
docker exec -it carton-kafka kafka-consumer-groups --describe \
  --bootstrap-server localhost:9092 \
  --group order-processors

# Вывод покажет:
# TOPIC           PARTITION  CURRENT-OFFSET  LOG-END-OFFSET  LAG  CONSUMER-ID
# orders.created  0          5               5               0    consumer-1
# orders.created  1          3               3               0    consumer-2
# orders.created  2          2               2               0    consumer-3
```

---

## 🎓 Практика #6: Kafka + Go Code

### Задача: Реализуй Order Event Producer

```go
package main

import (
	"encoding/json"
	"log"
	"time"
	
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type OrderCreatedEvent struct {
	OrderID   string    `json:"order_id"`
	UserID    string    `json:"user_id"`
	Amount    float64   `json:"amount"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

func main() {
	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer producer.Close()
	
	// Создай событие
	event := OrderCreatedEvent{
		OrderID:   "ORD-12345",
		UserID:    "USR-67890",
		Amount:    1299.99,
		Status:    "pending",
		CreatedAt: time.Now(),
	}
	
	// Сериализуй в JSON
	value, _ := json.Marshal(event)
	
	// Отправь в Kafka
	topic := "orders.created"
	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		Key:   []byte(event.OrderID),
		Value: value,
	}
	
	err = producer.Produce(msg, nil)
	if err != nil {
		log.Fatal(err)
	}
	
	// Жди подтверждения
	e := <-producer.Events()
	m := e.(*kafka.Message)
	
	if m.TopicPartition.Error != nil {
		log.Printf("Failed to deliver message: %v\n", m.TopicPartition.Error)
	} else {
		log.Printf("Delivered to partition %d at offset %d\n",
			m.TopicPartition.Partition, m.TopicPartition.Offset)
	}
}
```

**Задание**: Добавь retry логику и обработку ошибок.

---

## 🎓 Практика #7: Kubernetes Autoscaling

### Задача: Настрой HPA для Kafka Consumers

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: order-consumer-hpa
  namespace: carton-dream
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: order-consumer
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Pods
    pods:
      metric:
        name: kafka_consumer_lag
      target:
        type: AverageValue
        averageValue: "1000"
```

---

## 📊 Мониторинг Kafka

### Grafana Dashboards

```bash
# Открой Grafana (если запущен через docker-compose)
open http://localhost:3000

# Импортируй готовый dashboard для Kafka:
# Dashboard ID: 7589 (Kafka Overview)
```

### Prometheus Metrics

Kafka JMX Exporter уже настроен в `docker-compose.yml`. Метрики доступны:

```bash
curl http://localhost:9090/metrics | grep kafka
```

**Ключевые метрики**:
- `kafka_server_brokertopicmetrics_messagesinpersec` — incoming messages
- `kafka_server_brokertopicmetrics_bytesoutpersec` — outgoing bytes
- `kafka_controller_kafkacontroller_activecontrollercount` — active controller

---

## 🎯 Финальные задания

### 1. Реализуй Order Saga с Kafka
- Orders Service создает заказ → публикует `order.created`
- Payment Service слушает → обрабатывает → публикует `payment.completed`
- Notification Service слушает → отправляет email

### 2. Добавь Dead Letter Queue (DLQ)
- Если обработка сообщения упала 3 раза → отправь в `orders.created.dlq`
- Реализуй retry логику с exponential backoff

### 3. Настрой Kafka Streams
- Создай агрегацию: подсчитывай количество заказов по статусам за последний час
- Используй Kafka Streams API

### 4. Деплой в Production K8s
- Используй Helm для деплоя
- Настрой Ingress для доступа снаружи
- Добавь monitoring (Prometheus + Grafana)
- Настрой alerts для Kafka lag

---

## 🗡️ Мастер Мусаси говорит:

> _"Kafka — это река сообщений. Она течет непрерывно. Consumer — это рыбак, который ловит сообщения. Партиции — это разные участки реки. Consumer Group — это команда рыбаков, каждый на своем участке. Твоя задача — не дать реке выйти из берегов (lag). Начни с простого: отправь одно сообщение. Прими одно сообщение. Затем — тысячи."_

---

**Запуск проекта**: `./start.sh` или `docker-compose up -d`  
**Kafka UI**: http://localhost:8080  
**RabbitMQ UI**: http://localhost:15672  
**Admin API**: http://localhost:8082
