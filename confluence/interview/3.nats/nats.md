📡 3. СОБЫТИЙНАЯ АРХИТЕКТУРА (NATS, JETSTREAM)

❓ Что такое JetStream и чем отличается от Core NATS?
    Ответ:

    Core NATS — простая доставка "at most once", сообщения не сохраняются. Если потребитель оффлайн — сообщение потеряно.

    JetStream — добавляет персистентность, гарантии доставки, replay, Consumer Groups.

    В Event Horizon я использую JetStream для хранения всех событий лидерборда и магазина.

    Код: services/nats-hub/main.go — настройка Stream:

    go
    _, err = js.AddStream(&nats.StreamConfig{
        Name:     "EVENTS",
        Subjects: []string{"score.updated", "user.registered", "shop.purchased"},
        Storage:  nats.FileStorage,
        MaxAge:   7 * 24 * time.Hour,
    })


❓ Как NATS гарантирует, что сообщение не потеряется?
    Ответ: Когда потребитель забирает сообщение, он отправляет ack. Если ack не пришел в течение AckWait, сообщение переотправляется.

    Как это работает:

    Сообщение сохраняется на диске (JetStream)

    Потребитель получает сообщение, обрабатывает

    Потребитель отправляет ack

    NATS удаляет сообщение

    В коде Billing:

    go
    sub, err := js.Subscribe("score.updated", func(msg *nats.Msg) {
        // Обработка
        if err := process(msg); err == nil {
            msg.Ack() // Подтверждение
        }
    }, nats.AckWait(30*time.Second))


❓ Что такое durable subscription?
    Ответ: Это подписка, которая сохраняет свою позицию в Stream между перезапусками. Когда потребитель перезапускается, он продолжает с того места, где остановился.

    Под капотом:

    Создается Consumer с именем

    NATS сохраняет последний обработанный offset

    При перезапуске — восстанавливается

    В Event Horizon:

    go
    sub, err := js.Subscribe("score.updated", handler, 
        nats.Durable("billing-durable"), // ← durable имя
        nats.DeliverAll(), // Начать с первого
    )


❓ Как обрабатываешь дубликаты сообщений?
    Ответ: Использую idempotency keys (идентификаторы идемпотентности).

    Как работает:

    Каждое событие содержит reference_id (уникальный)

    В транзакции проверяем, не обрабатывали ли уже это событие

    Если reference_id уже есть — игнорируем

    В коде Billing:

    go
    // Добавляем баланс с проверкой reference_id
    _, err := s.pgRepo.AddBalance(ctx, userID, currency, amount, reason, referenceID)
    // Если reference_id уже существует — БД вернет ошибку уникальности


❓ Как бы реализовал Dead Letter Queue (DLQ)?
    Ответ: В NATS нет встроенной DLQ, но можно сделать самому:

    Создать отдельный Stream EVENTS_DLQ

    При ошибке после N-попыток публиковать в DLQ

    Мониторить DLQ в Grafana

    Реализация:

    go
    func processWithRetry(msg *nats.Msg) {
        err := process(msg)
        if err != nil && msg.Metadata().NumDelivered > 3 {
            // Отправляем в DLQ
            js.Publish("dlq.score.updated", msg.Data)
            msg.Ack()
            return
        }
        // Повторная попытка
        msg.Nak()
    }


❓ Ack vs AckWait?
    Ответ:

    Ack — немедленное подтверждение. NATS удаляет сообщение.

    Ack с AckWait — потребитель может взять паузу и подтвердить позже (до истечения AckWait).

    Пример:

    go
    // AckWait 30 секунд на обработку
    nats.AckWait(30*time.Second)

    // Если обработка займет больше — отправить Nak или Ack


❓ Что если потребитель умер после 10 переотправок?
    Ответ: NATS добавит сообщение в метаданные счетчик NumDelivered. Если он превысит лимит, сообщение будет удалено или отправлено в DLQ (если настроено).

    go
    // Проверяем счетчик доставок
    if msg.Metadata().NumDelivered > 5 {
        // Отправляем в DLQ
        js.Publish("dlq.score.updated", msg.Data)
        msg.Ack()
    }


❓ Почему subject'ы, а не топики?
    Ответ: Subject'ы в NATS поддерживают wildcards, что дает гибкость:

    score.updated — точное совпадение

    score.* — все игры (score.hexagon, score.flappy)

    score.> — все вложенные subject'ы

    В Event Horizon:

    go
    // Подписка на все игры
    js.Subscribe("score.*", handler)

    // Публикация для конкретной игры
    js.Publish("score.hexagon", data)


❓ Как гарантируешь порядок сообщений с несколькими потребителями?
    Ответ: В NATS порядок гарантируется только для одного потребителя в одном Stream. Если несколько потребителей — порядок не гарантируется.

    Решение:

    Использовать один Consumer Group с одним потребителем

    Использовать ключ партиционирования (user_id)

    В Event Horizon: У Billing один потребитель, поэтому порядок сохраняется.


❓ Что если один брокер упадет?
    Ответ: NATS кластер (3 ноды) автоматически переизбирает лидера. Если упал лидер — другой брокер становится новым лидером. Потребители переподключаются автоматически.

    Как проверить:

    bash
    # Проверяем текущего лидера
    curl -s http://localhost:8222/varz | jq '.leader'
    В Event Horizon: У нас кластер из 3 нод, поэтому отказа нет.