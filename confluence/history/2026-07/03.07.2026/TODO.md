# PROFILING

_ "net/http/pprof"


🔧 Что даст pprof
После добавления и пересборки ты получишь эндпоинты:

Эндпоинт	Описание
/debug/pprof/	Главная страница со списком профилей
/debug/pprof/profile?seconds=30	CPU профиль (30 секунд)
/debug/pprof/heap	Профиль памяти (heap)
/debug/pprof/goroutine	Стек всех горутин
/debug/pprof/threadcreate	Профиль создания потоков
/debug/pprof/block	Блокировки (mutex)
/debug/pprof/mutex	Конфликты мьютексов
🚀 Как проверить
bash
# 1. Пересобрать и перезапустить сервисы
make deploy

# 2. Проверить pprof для Game
curl -s "http://localhost:9092/debug/pprof/" | head -20

# 3. Снять 30-секундный CPU профиль
curl -s "http://localhost:9092/debug/pprof/profile?seconds=10" > /tmp/game_cpu.prof

# 4. Посмотреть горутины
curl -s "http://localhost:9092/debug/pprof/goroutine?debug=1" | head -50

### 7. pprof профилирование
- [ ] Добавить `import _ "net/http/pprof"` во все сервисы
- [ ] Проверить доступность `/debug/pprof/`
- [ ] Написать скрипт для автоматического сбора профилей
- [ ] Добавить в Grafana (через pyroscope/parca)