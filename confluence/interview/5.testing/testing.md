🧪 5. ТЕСТИРОВАНИЕ

❓ Какие метрики смотрел в K6? Что было узким местом?
Ответ: В первую очередь я смотрел:

p95 latency — чтобы понять, как быстро обрабатываются 95% запросов.

Error rate — сколько запросов падает.

RPS — пропускная способность.

Узкое место: Billing без Redis не выдерживал 500 VUs — ошибки были 28%, а p95 latency достигала 26 секунд. После добавления кеширования p95 упала до 200 мс.

Код: deployments/k6/loadtest.js — тест покрывает все эндпоинты.



❓ Как тестируешь NATS-консьюмеров в Go?
Ответ: Использую реальный NATS в тестах, запущенный через Docker.

go
func TestBillingSubscriber(t *testing.T) {
    // Поднимаем NATS
    nc, _ := nats.Connect(nats.DefaultURL)
    js, _ := nc.JetStream()
    
    // Публикуем тестовое событие
    js.Publish("score.updated", []byte(`{"user_id":"test","score":100}`))
    
    // Проверяем, что валюту начислили
    balance, _ := pgRepo.GetBalance(ctx, "test", "tickets")
    assert.Equal(t, 10, balance)
}
Моки: Для unit-тестов использую nats-server в testing режиме.



❓ Что такое testcontainers?
Ответ: Библиотека, которая поднимает реальный PostgreSQL в Docker для интеграционных тестов.

Пример:

go
func TestPostgresRepo(t *testing.T) {
    container, _ := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
        ContainerRequest: testcontainers.ContainerRequest{
            Image: "postgres:16-alpine",
            Env: map[string]string{"POSTGRES_PASSWORD": "test"},
        },
    })
    defer container.Terminate(ctx)
    
    // Подключаемся и тестируем
}
В Event Horizon: Использую для тестирования репозиториев.





❓ Как тестировать WebSocket под нагрузкой 10 000 пользователей?
Ответ: В K6 есть поддержка WebSocket:

javascript
const ws = new WebSocket('ws://localhost:8079/ws/leaderboard');
ws.on('message', (data) => {
    // Проверяем, что приходят обновления
});

ws.send(JSON.stringify({ command: 'subscribe', game: 'hexagon' }));
Что проверяю:

Успешность подключения

Задержка доставки сообщений

Ошибки при массовых подключениях



❓ Как проверяешь обработку context и таймаутов?
Ответ: Использую context.WithTimeout в тестах:

go
func TestTimeout(t *testing.T) {
    ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
    defer cancel()
    
    _, err := service.GetBalance(ctx, "test-user")
    assert.ErrorIs(t, err, context.DeadlineExceeded)
}
В продакшене: Все gRPC-запросы имеют таймаут 5 секунд.



❓ Что такое table-driven tests?
Ответ: Это подход, где тест-кейсы объединены в таблицу:

go
func TestAddBalance(t *testing.T) {
    tests := []struct {
        name     string
        amount   int
        expected int
    }{
        {"positive", 10, 10},
        {"zero", 0, 0},
        {"negative", -5, -5},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := AddBalance(tt.amount)
            assert.Equal(t, tt.expected, result)
        })
    }
}
Почему хорошо: Легко добавлять новые кейсы, читаемо.



❓ Как тестировать outbox-паттерн с 10 сервисами?
Ответ: Тестирую один сервис за раз с помощью интеграционных тестов.

Запускаю тестовый NATS и PostgreSQL

Публикую событие через outbox

Проверяю, что событие ушло в очередь

Проверяю, что потребитель обработал

Для 10 сервисов: Тестирую каждый сервис отдельно, не эмуляцию всех 10.



❓ Если 90% покрытие, а баг есть — почему?
Ответ: Потому что покрытие не гарантирует корректность. Баг может быть:

В интеграции между сервисами (тесты не покрывают)

В конфигурации (не тестируется)

В данных (не все кейсы покрыты)

Как исправить: Добавить интеграционные тесты с реальными зависимостями.



❓ Какие метрики собираешь в K6?
Ответ:

http_req_duration — p50, p90, p95, p99

http_req_failed — процент ошибок

iterations — сколько итераций выполнено

data_sent / data_received — трафик

Когда система не выдерживает: p95 > 1s или error rate > 1%.



❓ Как тестировать graceful shutdown?
Ответ:

Отправить сигнал SIGTERM

Проверить, что сервис не принимает новые запросы

Проверить, что текущие запросы завершились

Проверить, что БД обновлена

В коде:

go
func main() {
    // Обработка SIGTERM
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGTERM)
    <-quit
    
    // Graceful shutdown
    grpcServer.GracefulStop()
    nc.Drain()
}

Что должно успеть:

Доиграть транзакции

Отправить ack в NATS

Сохранить данные в БД