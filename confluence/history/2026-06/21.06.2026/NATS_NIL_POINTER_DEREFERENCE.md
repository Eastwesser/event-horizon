НАСТОЯЩАЯ ПРОБЛЕМА!

Смотри:

Gateway упал с паникой — nil pointer dereference в NATS.
Balancer скомпилирован со старыми бекендами — он всё ещё ищет 127.0.0.1:8080, а не gateway:8080.
Проблема 1: Gateway падает из-за NATS

Gateway пытается вызвать nc.Drain() на nil-объекте, потому что NATS не подключился.

Решение: Правим код gateway:

bash
cd /home/denismatveev/event_horizon/services/gateway

# В функции main, после nc, err := nats.Connect(...)
# Нужно добавить проверку на nil перед Drain()
В cmd/main.go найди строки:

go
nc, err := nats.Connect(cfg.NATSUrl)
if err != nil {
    log.Printf("⚠️ Failed to connect to NATS: %v (WebSocket будет недоступен)", err)
}
defer nc.Drain()  // ← ЭТО ЛОМАЕТ, если nc == nil
Замени на:

go
nc, err := nats.Connect(cfg.NATSUrl)
if err != nil {
    log.Printf("⚠️ Failed to connect to NATS: %v (WebSocket будет недоступен)", err)
    // Не делаем defer nc.Drain() если nc == nil
} else {
    defer nc.Drain()
}
Или проще — временно закомментируй defer nc.Drain() и js, err := nc.JetStream() если nc == nil.

