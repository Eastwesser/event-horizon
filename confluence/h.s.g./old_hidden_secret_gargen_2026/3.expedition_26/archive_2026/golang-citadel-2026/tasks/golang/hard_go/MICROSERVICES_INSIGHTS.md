# 🏗️ Microservices Insights — Паттерны микросервисов для Senior

**Цель:** Разобрать продвинутые паттерны микросервисной архитектуры.

---

## 📋 Список паттернов

1. **Lazy Loading** — Отложенная загрузка данных
2. **DLQ** (Dead Letter Queue) — Очередь недоставленных сообщений
3. **SAGA** — Распределённые транзакции
4. **Dispatcher** — Диспетчеризация запросов
5. **Batch Processing** — Пакетная обработка
6. **Pub/Sub** — Публикация/подписка

---

## 🎯 Детальный разбор

### 1. Lazy Loading — Отложенная загрузка

**Что это:**
- Загрузка данных **только когда они нужны**
- Оптимизация памяти и производительности
- Избегаем N+1 проблемы

**Проблема (N+1 Query):**
```go
// ❌ ПЛОХО: N+1 запросов к БД
func GetUsersWithOrders() []UserWithOrders {
    users := db.FindAllUsers()  // 1 запрос
    
    for _, user := range users {
        user.Orders = db.FindOrdersByUser(user.ID)  // N запросов!
    }
    
    return users
}
```

**Решение 1: Eager Loading (Жадная загрузка):**
```go
// ✅ ХОРОШО: 2 запроса
func GetUsersWithOrders() []UserWithOrders {
    users := db.FindAllUsers()  // 1 запрос
    
    userIDs := extractIDs(users)
    orders := db.FindOrdersByUserIDs(userIDs)  // 1 запрос
    
    // Группируем заказы по user_id
    ordersByUser := groupByUserID(orders)
    
    for i := range users {
        users[i].Orders = ordersByUser[users[i].ID]
    }
    
    return users
}
```

**Решение 2: Lazy Loading (через Proxy):**
```go
type LazyOrders struct {
    userID int
    db     *Database
    loaded bool
    orders []Order
}

func (lo *LazyOrders) Get() []Order {
    if !lo.loaded {
        lo.orders = lo.db.FindOrdersByUser(lo.userID)
        lo.loaded = true
    }
    return lo.orders
}

type User struct {
    ID     int
    Name   string
    Orders *LazyOrders  // Загрузятся только при вызове Get()
}
```

**Решение 3: DataLoader (Facebook pattern):**
```go
type DataLoader struct {
    batchFn   func([]int) []Order
    cache     map[int][]Order
    batch     []int
    mu        sync.Mutex
    waitGroup sync.WaitGroup
}

func (dl *DataLoader) Load(userID int) []Order {
    dl.mu.Lock()
    
    if orders, ok := dl.cache[userID]; ok {
        dl.mu.Unlock()
        return orders
    }
    
    dl.batch = append(dl.batch, userID)
    dl.waitGroup.Add(1)
    dl.mu.Unlock()
    
    // Trigger batch execution
    go dl.executeBatch()
    
    dl.waitGroup.Wait()
    return dl.cache[userID]
}

func (dl *DataLoader) executeBatch() {
    time.Sleep(10 * time.Millisecond)  // Collect batch
    
    dl.mu.Lock()
    batch := dl.batch
    dl.batch = nil
    dl.mu.Unlock()
    
    if len(batch) == 0 {
        return
    }
    
    orders := dl.batchFn(batch)
    
    dl.mu.Lock()
    for _, order := range orders {
        dl.cache[order.UserID] = append(dl.cache[order.UserID], order)
    }
    dl.mu.Unlock()
    
    for range batch {
        dl.waitGroup.Done()
    }
}
```

**Use case:**
- GraphQL queries
- REST API с вложенными ресурсами
- ORM (Hibernate, GORM)

---

### 2. DLQ (Dead Letter Queue)

**Что это:**
- Очередь для сообщений, которые **не удалось обработать**
- Retry mechanism с лимитом попыток
- Ручная обработка проблемных сообщений

**Архитектура:**
```
┌──────────┐          ┌──────────┐          ┌──────────┐
│ Producer │ ──────>  │  Queue   │ ──────>  │ Consumer │
└──────────┘          └──────────┘          └────┬─────┘
                                                  │
                                            Failed (3 retries)
                                                  │
                                                  ▼
                                           ┌──────────┐
                                           │   DLQ    │
                                           └──────────┘
                                                  │
                                            Manual Review
```

**Реализация (RabbitMQ):**
```go
// 1. Объявляем главную очередь с DLQ
channel.QueueDeclare(
    "orders",      // name
    true,          // durable
    false,         // auto-delete
    false,         // exclusive
    false,         // no-wait
    amqp.Table{
        "x-dead-letter-exchange":    "dlx",
        "x-dead-letter-routing-key": "orders.dlq",
        "x-message-ttl":             60000,  // 60 seconds
    },
)

// 2. Объявляем DLQ
channel.QueueDeclare(
    "orders.dlq",  // name
    true,          // durable
    false,         // auto-delete
    false,         // exclusive
    false,         // no-wait
    nil,
)

// 3. Consumer с retry logic
func processMessage(msg amqp.Delivery) error {
    retryCount := getRetryCount(msg.Headers)
    
    if err := handleOrder(msg.Body); err != nil {
        if retryCount < 3 {
            // Retry (reject → requeue)
            msg.Headers["x-retry-count"] = retryCount + 1
            return msg.Nack(false, true)
        } else {
            // Send to DLQ
            log.Printf("Max retries reached, sending to DLQ: %s", err)
            return msg.Nack(false, false)  // no requeue → DLQ
        }
    }
    
    return msg.Ack(false)
}
```

**Реализация (Kafka):**
```go
type MessageProcessor struct {
    mainTopic string
    dlqTopic  string
    producer  *kafka.Producer
}

func (p *MessageProcessor) Process(msg *kafka.Message) {
    retryCount := getRetryCount(msg.Headers)
    
    if err := handleMessage(msg.Value); err != nil {
        if retryCount < 3 {
            // Retry
            msg.Headers = append(msg.Headers, kafka.Header{
                Key:   "retry-count",
                Value: []byte(fmt.Sprintf("%d", retryCount+1)),
            })
            p.producer.Produce(&kafka.Message{
                TopicPartition: kafka.TopicPartition{
                    Topic:     &p.mainTopic,
                    Partition: kafka.PartitionAny,
                },
                Value:   msg.Value,
                Headers: msg.Headers,
            }, nil)
        } else {
            // Send to DLQ
            p.producer.Produce(&kafka.Message{
                TopicPartition: kafka.TopicPartition{
                    Topic:     &p.dlqTopic,
                    Partition: kafka.PartitionAny,
                },
                Value: msg.Value,
                Headers: append(msg.Headers, kafka.Header{
                    Key:   "error",
                    Value: []byte(err.Error()),
                }),
            }, nil)
        }
    }
}
```

**Use case:**
- Обработка заказов (если платёж упал)
- Email/SMS отправка (если API недоступен)
- Интеграции с внешними системами

---

### 3. SAGA — Распределённые транзакции

**Что это:**
- Паттерн для **распределённых транзакций** между микросервисами
- Sequence of local transactions
- Каждый шаг имеет **compensating transaction** (откат)

**Типы SAGA:**

**1. Choreography (Хореография) — Event-driven:**
```
Order Service ────> Order Created Event
                           │
                           ├──> Payment Service ────> Payment Processed
                           │                                   │
                           │                          Payment Failed
                           │                                   │
                           └──> Inventory Service      Compensate: Refund
                                        │
                                 Stock Reserved
```

**Реализация:**
```go
// Order Service
func CreateOrder(order Order) error {
    if err := db.SaveOrder(order); err != nil {
        return err
    }
    
    // Публикуем событие
    publishEvent(OrderCreatedEvent{
        OrderID: order.ID,
        UserID:  order.UserID,
        Items:   order.Items,
        Total:   order.Total,
    })
    
    return nil
}

// Payment Service (подписчик)
func HandleOrderCreated(event OrderCreatedEvent) {
    if err := processPayment(event.UserID, event.Total); err != nil {
        // Payment failed → compensate
        publishEvent(PaymentFailedEvent{
            OrderID: event.OrderID,
            Reason:  err.Error(),
        })
        return
    }
    
    publishEvent(PaymentProcessedEvent{
        OrderID: event.OrderID,
    })
}

// Order Service (подписчик на PaymentFailed)
func HandlePaymentFailed(event PaymentFailedEvent) {
    // Откатываем заказ
    db.UpdateOrderStatus(event.OrderID, "cancelled")
}
```

**2. Orchestration (Оркестрация) — Coordinator:**
```
                    ┌─────────────┐
                    │ SAGA Manager│
                    └──────┬──────┘
                           │
         ┌─────────────────┼─────────────────┐
         │                 │                 │
         ▼                 ▼                 ▼
   Order Service    Payment Service   Inventory Service
```

**Реализация:**
```go
type SagaManager struct {
    orderSvc     *OrderService
    paymentSvc   *PaymentService
    inventorySvc *InventoryService
}

func (sm *SagaManager) CreateOrder(order Order) error {
    // Step 1: Create Order
    if err := sm.orderSvc.Create(order); err != nil {
        return err
    }
    
    // Step 2: Process Payment
    if err := sm.paymentSvc.Process(order.UserID, order.Total); err != nil {
        // Compensate: Cancel Order
        sm.orderSvc.Cancel(order.ID)
        return err
    }
    
    // Step 3: Reserve Inventory
    if err := sm.inventorySvc.Reserve(order.Items); err != nil {
        // Compensate: Refund + Cancel Order
        sm.paymentSvc.Refund(order.UserID, order.Total)
        sm.orderSvc.Cancel(order.ID)
        return err
    }
    
    // Success: Complete Order
    sm.orderSvc.Complete(order.ID)
    return nil
}
```

**Use case:**
- E-commerce (заказ → оплата → резерв товара)
- Booking systems (бронирование → оплата → подтверждение)
- Financial transactions

---

### 4. Dispatcher — Диспетчеризация запросов

**Что это:**
- Центральный компонент для **маршрутизации** запросов
- Load balancing между сервисами
- Service discovery

**Паттерны:**

**1. API Gateway:**
```go
type APIGateway struct {
    userSvc    *UserServiceClient
    orderSvc   *OrderServiceClient
    productSvc *ProductServiceClient
}

func (gw *APIGateway) HandleRequest(r *http.Request) {
    switch r.URL.Path {
    case "/api/users":
        gw.userSvc.Handle(r)
    case "/api/orders":
        gw.orderSvc.Handle(r)
    case "/api/products":
        gw.productSvc.Handle(r)
    default:
        http.Error(w, "Not Found", 404)
    }
}
```

**2. Command Dispatcher:**
```go
type CommandDispatcher struct {
    handlers map[string]CommandHandler
}

func (d *CommandDispatcher) Dispatch(cmd Command) error {
    handler, ok := d.handlers[cmd.Type()]
    if !ok {
        return fmt.Errorf("no handler for command: %s", cmd.Type())
    }
    
    return handler.Handle(cmd)
}

// Регистрация handlers
dispatcher := &CommandDispatcher{
    handlers: map[string]CommandHandler{
        "CreateOrder":  &CreateOrderHandler{},
        "CancelOrder":  &CancelOrderHandler{},
        "UpdateOrder":  &UpdateOrderHandler{},
    },
}

dispatcher.Dispatch(CreateOrderCommand{...})
```

**3. Event Dispatcher:**
```go
type EventDispatcher struct {
    subscribers map[string][]EventHandler
}

func (d *EventDispatcher) Subscribe(eventType string, handler EventHandler) {
    d.subscribers[eventType] = append(d.subscribers[eventType], handler)
}

func (d *EventDispatcher) Publish(event Event) {
    handlers := d.subscribers[event.Type()]
    
    for _, handler := range handlers {
        go handler.Handle(event)  // Асинхронно
    }
}
```

**Use case:**
- API Gateway (Kong, Nginx)
- Event-driven architecture
- CQRS (Command/Query routing)

---

### 5. Batch Processing — Пакетная обработка

**Что это:**
- Обработка **группы** сообщений/запросов за раз
- Оптимизация DB операций (bulk insert)
- Снижение network overhead

**Реализация:**
```go
type BatchProcessor struct {
    batch     []Message
    batchSize int
    ticker    *time.Ticker
    mu        sync.Mutex
}

func NewBatchProcessor(batchSize int, interval time.Duration) *BatchProcessor {
    bp := &BatchProcessor{
        batch:     make([]Message, 0, batchSize),
        batchSize: batchSize,
        ticker:    time.NewTicker(interval),
    }
    
    go bp.run()
    return bp
}

func (bp *BatchProcessor) Add(msg Message) {
    bp.mu.Lock()
    defer bp.mu.Unlock()
    
    bp.batch = append(bp.batch, msg)
    
    if len(bp.batch) >= bp.batchSize {
        bp.flush()
    }
}

func (bp *BatchProcessor) run() {
    for range bp.ticker.C {
        bp.mu.Lock()
        bp.flush()
        bp.mu.Unlock()
    }
}

func (bp *BatchProcessor) flush() {
    if len(bp.batch) == 0 {
        return
    }
    
    // Bulk insert в БД
    db.BulkInsert(bp.batch)
    
    bp.batch = bp.batch[:0]  // Clear
}
```

**Kafka Consumer с batch:**
```go
func consumeBatch(consumer *kafka.Consumer) {
    batch := make([]*kafka.Message, 0, 100)
    timeout := time.After(5 * time.Second)
    
    for {
        select {
        case msg := <-consumer.Messages():
            batch = append(batch, msg)
            
            if len(batch) >= 100 {
                processBatch(batch)
                batch = batch[:0]
            }
            
        case <-timeout:
            if len(batch) > 0 {
                processBatch(batch)
                batch = batch[:0]
            }
            timeout = time.After(5 * time.Second)
        }
    }
}

func processBatch(batch []*kafka.Message) {
    // Bulk insert
    db.BulkInsert(extractData(batch))
    
    // Commit offsets
    for _, msg := range batch {
        consumer.CommitMessage(msg)
    }
}
```

**Use case:**
- ETL pipelines
- Analytics (batch metrics)
- Email/SMS campaigns (bulk send)

---

### 6. Pub/Sub — Публикация/подписка

**Что это:**
- Асинхронная коммуникация между сервисами
- Publisher не знает о Subscribers
- Decoupling (слабая связанность)

**Архитектура:**
```
┌─────────┐          ┌───────────┐
│Publisher│──────────>│Event Bus  │
└─────────┘          │(Kafka/    │
                     │RabbitMQ)  │
                     └─────┬─────┘
                           │
          ┌────────────────┼────────────────┐
          │                │                │
          ▼                ▼                ▼
   ┌──────────┐     ┌──────────┐    ┌──────────┐
   │Subscriber│     │Subscriber│    │Subscriber│
   │    A     │     │    B     │    │    C     │
   └──────────┘     └──────────┘    └──────────┘
```

**Реализация (In-memory):**
```go
type EventBus struct {
    subscribers map[string][]chan Event
    mu          sync.RWMutex
}

func (eb *EventBus) Subscribe(eventType string) <-chan Event {
    eb.mu.Lock()
    defer eb.mu.Unlock()
    
    ch := make(chan Event, 10)
    eb.subscribers[eventType] = append(eb.subscribers[eventType], ch)
    return ch
}

func (eb *EventBus) Publish(event Event) {
    eb.mu.RLock()
    defer eb.mu.RUnlock()
    
    for _, ch := range eb.subscribers[event.Type()] {
        go func(c chan Event) {
            c <- event
        }(ch)
    }
}

// Использование
bus := &EventBus{subscribers: make(map[string][]chan Event)}

// Service A (subscriber)
orderCreated := bus.Subscribe("OrderCreated")
go func() {
    for event := range orderCreated {
        log.Printf("Service A: Processing %+v", event)
    }
}()

// Service B (publisher)
bus.Publish(OrderCreatedEvent{OrderID: 123})
```

**Реализация (Kafka):**
```go
// Publisher
func publishEvent(producer *kafka.Producer, topic string, event Event) {
    data, _ := json.Marshal(event)
    
    producer.Produce(&kafka.Message{
        TopicPartition: kafka.TopicPartition{
            Topic:     &topic,
            Partition: kafka.PartitionAny,
        },
        Value: data,
    }, nil)
}

// Subscriber
func subscribe(consumer *kafka.Consumer, topic string, handler func(Event)) {
    consumer.Subscribe([]string{topic}, nil)
    
    for {
        msg, _ := consumer.ReadMessage(-1)
        
        var event Event
        json.Unmarshal(msg.Value, &event)
        
        handler(event)
    }
}
```

**Use case:**
- Event-driven architecture
- Microservices communication
- Real-time notifications

---

## 🎯 Советы для собеса

### 1. Обсуждай trade-offs
- **SAGA Choreography vs Orchestration:**
  - Choreography → decoupled, но сложнее отладка
  - Orchestration → проще отладка, но single point of failure

- **Sync vs Async:**
  - Sync (HTTP) → простота, но tight coupling
  - Async (Kafka) → scalability, но eventual consistency

### 2. Знай CAP теорему
- **Consistency:** Все видят одни данные
- **Availability:** Система всегда отвечает
- **Partition tolerance:** Работа при network partition

**Невозможно иметь все 3 одновременно!**

### 3. Практические вопросы
- "Как обеспечить exactly-once delivery в Kafka?"
- "Как rollback SAGA при ошибке?"
- "Как избежать cascading failures?" (Circuit Breaker!)

---

**Удачи на Senior собесе!** 🏗️
