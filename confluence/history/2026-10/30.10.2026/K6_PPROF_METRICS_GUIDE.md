#  ГАЙД: ЗАМЕР ПРОИЗВОДИТЕЛЬНОСТИ EVENT HORIZON (k6 + pprof)

**Дата:** 30 октября 2026  
**Цель:** Собрать CPU-профиль под нагрузкой и визуализировать его в PDF.

---

## 📋 ПРЕДУСЛОВИЯ

1. **Стек запущен:**
   ```bash
   make status
Все контейнеры должны быть в состоянии Up или healthy.

k6 установлен:

bash
k6 version
Если нет: sudo pacman -S k6 (Arch Linux).

Graphviz установлен (для генерации PDF):

bash
dot -V
Если нет: sudo pacman -S graphviz.

pprof включен в gateway:
В services/gateway/cmd/main.go должен быть:

go
import _ "net/http/pprof"
И запущен HTTP-сервер на :6060:

go
go func() {
    log.Println("pprof server listening on :6060")
    http.ListenAndServe(":6060", nil)
}()
Порт 6060 проброшен в deployments/docker-compose.cluster.yml:

yaml
gateway:
  ports:
    - "8081:8080"
    - "6060:6060"   # pprof
🚀 ШАГ 1: ЗАПУСТИТЬ НАГРУЗКУ (K6)
Есть два способа:

Вариант А: Через Makefile (правильный путь)
bash
make test-k6
Это запустит deployments/k6/browse.js против балансировщика на :8079 с 20 VU на 30 секунд.

Важно: нужен scripts/.env.seed.admin с логином/паролем админа. Если его нет:

bash
cp scripts/.env.seed.admin.example scripts/.env.seed.admin
# Отредактируй: SEED_ADMIN_EMAIL и SEED_ADMIN_PASSWORD
Вариант Б: Вручную (если хочешь свой скрипт)
bash
BASE_URL=http://localhost:8079 k6 run scripts/loadtest.js
Ключевой момент: BASE_URL должен быть :8079 (балансировщик), а не :8080 (внутренний порт gateway).

🚀 ШАГ 2: СОБРАТЬ CPU-ПРОФИЛЬ (ПОКА K6 РАБОТАЕТ)
Открой второй терминал и выполни:

bash
curl -o cpu.out "http://localhost:6060/debug/pprof/profile?seconds=30"
Что это делает:

Подключается к pprof-серверу gateway.

Собирает CPU-профиль за 30 секунд.

Сохраняет в cpu.out.

Результат: файл cpu.out (~30–50 КБ).

🚀 ШАГ 3: КОНВЕРТИРОВАТЬ В PDF
bash
go tool pprof -pdf cpu.out > cpu.pdf
Результат: файл cpu.pdf с графом вызовов.

Если ошибка dot not found — установи Graphviz:

bash
sudo pacman -S graphviz
🚀 ШАГ 4: ПЕРЕМЕСТИТЬ В АРХИВ
Создай папку для профилей:

bash
mkdir -p confluence/history/$(date +%Y-%m)/$(date +%d.%m.%Y)/profiles
mv cpu.pdf confluence/history/$(date +%Y-%m)/$(date +%d.%m.%Y)/profiles/
mv cpu.out confluence/history/$(date +%Y-%m)/$(date +%d.%m.%Y)/profiles/
Важно: cpu.out — бинарный файл, он не попадет в Git (в .gitignore есть *.out). cpu.pdf — документ, его можно коммитить (если хочешь).

🚀 ШАГ 5: АНАЛИЗ ПРОФИЛЯ
Открой cpu.pdf и посмотри на топ функций (самые «горячие»).

На что обратить внимание:

Функция	Что это	Что делать
runtime.mallocgc	Аллокации памяти	Проверь, нет ли лишних аллокаций в hot path. Используй sync.Pool.
runtime.gcBgMarkWorker	Сборка мусора	Если GC работает часто — проверь GOGC и уменьши аллокации.
net/http.(*conn).serve	HTTP-обработка	Проверь middleware, не делают ли они лишних аллокаций.
google.golang.org/grpc	gRPC-вызовы	Проверь, нет ли лишних вызовов или больших payload.
database/sql	Запросы к БД	Проверь индексы и пул соединений.
🚀 ШАГ 6: ИНТЕРПРЕТАЦИЯ РЕЗУЛЬТАТОВ K6
Смотри на ключевые метрики:

Метрика	Что значит	Норма
http_reqs	RPS (запросов в секунду)	Зависит от нагрузки
http_req_failed	Процент ошибок	< 1%
p(95) latency	95% запросов быстрее этого	< 500 мс (для API)
iterations	Количество итераций	Зависит от VU
Пример хорошего результата (30.10.2026):

text
http_reqs: 1299 (40.8 RPS)
http_req_failed: 0.00%
p(95): 539 мс
