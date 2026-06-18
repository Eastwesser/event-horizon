# Jaeger — Успешная интеграция трейсинга
**Дата:** 18 июня 2026  
**Статус:** ✅ ГОТОВО

---

## 📌 Цель
Настроить распределённый трейсинг для EventHorizon через OpenTelemetry + Jaeger, чтобы видеть полный путь запроса через все микросервисы.

---

## 🧩 Ситуация (Situation)
- EventHorizon состоит из 5 микросервисов: Auth, Game, Billing, Leaderboard, Gateway
- Для отладки и понимания производительности нужен распределённый трейсинг
- Ранее Jaeger был запущен, но сервисы не отправляли трейсы

---

## ⚔️ Задача (Task)
1. Добавить OpenTelemetry во все 5 сервисов
2. Настроить экспорт трейсов в Jaeger (OTLP gRPC)
3. Убедиться, что трейсы отображаются в Jaeger UI
4. Задокументировать решение

---

## 🛠️ Действия (Action)

### 1. Добавление OpenTelemetry в сервисы
- Во все `main.go` добавлена функция `initTracer()`
- В gRPC-сервисы добавлены интерсепторы:
  ```go
  grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor())
  grpc.StreamInterceptor(otelgrpc.StreamServerInterceptor())
В Gateway (Gin) добавлена мидлвара: otelgin.Middleware("gateway")
2. Настройка эндпоинта

В initTracer() добавлено чтение переменной окружения:
go
endpoint := os.Getenv("JAEGER_ENDPOINT")
if endpoint == "" {
    endpoint = "localhost:4317"
}
В ~/.bashrc и .env добавлено:
bash
export JAEGER_ENDPOINT=172.17.0.1:4317
3. Проброс порта в Docker

В docker-compose.cluster.yml добавлен порт 4317 для Jaeger:
yaml
jaeger:
  ports:
    - "16686:16686"   # UI
    - "4318:4318"     # OTLP HTTP
    - "4317:4317"     # OTLP gRPC
4. Диагностика и отладка

В initTracer() добавлены логи:
text
🔄 Initializing Jaeger tracer with endpoint: 172.17.0.1:4317
✅ Jaeger exporter created
✅ Jaeger tracer initialized
Ошибка connection refused устранена пробросом порта
✅ Результат (Result)

Итоговый стек

Компонент	Статус	Детали
Auth	✅	Трейсы отправляются в Jaeger
Billing	✅	Трейсы отправляются в Jaeger
Game	✅	Трейсы отправляются в Jaeger
Gateway	✅	Gin-мидлвара активна
Leaderboard	✅	Трейсы отправляются в Jaeger
Jaeger UI	✅	Доступен на http://localhost:16686
Конфигурация

Эндпоинт: 172.17.0.1:4317 (через переменную JAEGER_ENDPOINT)
Порт Jaeger gRPC: 4317
Порт Jaeger HTTP: 4318
Порт Jaeger UI: 16686
Логи успешной инициализации

text
2026/06/18 19:07:21 🔄 Initializing Jaeger tracer with endpoint: 172.17.0.1:4317
2026/06/18 19:07:21 ✅ Jaeger exporter created
2026/06/18 19:07:21 ✅ Jaeger tracer initialized
2026/06/18 19:07:21 📊 Metrics endpoint: http://localhost:9091/metrics
2026/06/18 19:07:21 🔐 Auth service listening on :50051
Jaeger UI

Сервисы: auth, gateway, game, billing, leaderboard (появляются по мере генерации трафика)
Трейсы: видны полные пути запросов с таймингами

📝 Выводы (STAR-анализ)

Situation

EventHorizon — распределённая система из 5 микросервисов. Для наблюдения за производительностью и отладки требовался распределённый трейсинг.

Task

Настроить сбор и визуализацию трейсов через OpenTelemetry + Jaeger для всех сервисов.

Action

Внедрил OpenTelemetry во все 5 сервисов (gRPC-интерсепторы + Gin-мидлвара)
Настроил чтение эндпоинта из переменной окружения JAEGER_ENDPOINT
Добавил проброс порта 4317 в Docker Compose
Добавил логирование для диагностики подключения

Result

✅ Все сервисы успешно отправляют трейсы в Jaeger
✅ Jaeger UI отображает полные трассы запросов
✅ Ошибки connection refused устранены
✅ Система готова к нагрузочному тестированию с k6

🚀 Следующие шаги

Нагрузочное тестирование (k6) — прогнать 4 игры под нагрузкой
Сравнить — что показывают трейсы в Jaeger vs метрики в Grafana
Настроить алерты по latency (SLO)

Дата завершения: 18 июня 2026, 19:07
Автор: s1ntezc0der и Хохо-2
Статус: ✅ Готово к нагрузочному тестированию