pprof — это швейцарский нож для Go-разработчика. В отличие от метрик (которые показывают "что"), pprof показывает "почему" — где тормозит код, где утечка памяти, где блокировки.

Добавим его в наш план.

🧠 Что такое pprof и зачем он нам
pprof — это встроенный в Go профилировщик. Он даёт:

CPU профиль — где процессор тратит время.

Heap профиль — кто выделяет память (поиск утечек).

Goroutine профиль — сколько горутин, где они висят.

Mutex профиль — кто держит мьютексы.

Block профиль — где горутины блокируются.

В отличие от Prometheus (который даёт агрегированные метрики), pprof даёт детальный срез состояния приложения в момент времени.

✅ Что у нас уже есть и чего не хватает
Сейчас в проекте:
✅ Метрики через Prometheus (количество запросов, память, горутины).

❌ Нет pprof эндпоинтов (нигде не включён).

Надо добавить:
Включить net/http/pprof в Gateway и всех сервисах.

Настроить сбор профилей в production (безопасно).

Интегрировать с Grafana (через pyroscope или parca).

🛠️ Как добавить pprof в проект
1. Добавить import в main.go каждого сервиса
go
import _ "net/http/pprof"
2. Включить pprof на отдельном порту (или на том же, где метрики)
go
// В main.go, в том же goroutine, где метрики:
go func() {
    log.Printf("📊 pprof endpoint: http://0.0.0.0:%s/debug/pprof/", cfg.MetricsPort)
    if err := http.ListenAndServe(":"+cfg.MetricsPort, nil); err != nil {
        log.Printf("pprof server error: %v", err)
    }
}()
Важно: если ты уже используешь http.Handle("/metrics", promhttp.Handler()), то pprof добавится автоматически на тот же порт (через import _ "net/http/pprof").

🔧 Как собирать профили
Вручную (для отладки)
bash
# CPU профиль (30 секунд)
curl -s http://localhost:9095/debug/pprof/profile?seconds=30 > cpu.prof

# Heap профиль (память)
curl -s http://localhost:9095/debug/pprof/heap > heap.prof

# Goroutine профиль
curl -s http://localhost:9095/debug/pprof/goroutine > goroutine.prof

# Mutex профиль
curl -s http://localhost:9095/debug/pprof/mutex > mutex.prof
Анализ профиля
bash
# Интерактивный режим (top, list, web)
go tool pprof -http=:8081 cpu.prof
Автоматический сбор (через скрипт)
Добавим в metrics_collector.sh:

bash
# Сбор профилей раз в час (если есть нагрузка)
collect_profiles() {
    local port=$1
    local service=$2
    local timestamp=$(date +"%Y%m%d_%H%M%S")
    curl -s "http://localhost:${port}/debug/pprof/profile?seconds=30" > "${OUTPUT_DIR}/${service}_cpu_${timestamp}.prof"
    curl -s "http://localhost:${port}/debug/pprof/heap" > "${OUTPUT_DIR}/${service}_heap_${timestamp}.prof"
    curl -s "http://localhost:${port}/debug/pprof/goroutine" > "${OUTPUT_DIR}/${service}_goroutine_${timestamp}.prof"
}
🛡️ Безопасность в Production
pprof может быть опасен (отдаёт много данных, может нагружать CPU). Поэтому:

Не открывай pprof наружу — только через внутреннюю сеть.

Используй отдельный порт (или включай только по запросу).

Добавь аутентификацию (через httpauth или reverse proxy).

Включай pprof только на стейджинге (в проде — только по требованию).

📋 TODO: pprof интеграция
Добавить import _ "net/http/pprof" в services/*/cmd/main.go

Убедиться, что pprof работает на порту метрик (например, 9095/debug/pprof/)

Обновить metrics_collector.sh для сбора профилей

Настроить автоматический сбор профилей при высоком CPU/памяти

Добавить документацию по анализу профилей в confluence/

🔮 Дальнейшие планы
Pyroscope / Parca — непрерывный профилировщик (вместо ручного сбора).

Интеграция с Grafana — чтобы смотреть профили прямо в дашбордах.

Автоматический анализ — alert при утечке памяти или росте горутин.

