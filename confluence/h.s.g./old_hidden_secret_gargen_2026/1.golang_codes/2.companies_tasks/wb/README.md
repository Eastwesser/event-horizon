# WB — TechnoSchool haven wrapper

Тренировка под Wildberries через репозиторий Техношколы.

## Как гонять

```bash
export GOWORK=off
cd wb_technoschool-main/level_1/task_N   # или level_2/task_N
go run sol.go
# гонки (где уместно):
go run -race sol.go
```

Полное оглавление уровней: [`wb_technoschool-main/README.md`](wb_technoschool-main/README.md).

## Priority (~8) для middle+ (concurrency / sync / HTTP)

| Задача | Почему на собесе |
|--------|------------------|
| `level_1/task_2` | параллельные квадраты — базовые горутины |
| `level_1/task_3` | producer + N workers из канала |
| `level_1/task_4` | graceful shutdown по SIGINT (context) |
| `level_1/task_6` | способы стопа горутины |
| `level_1/task_7` | concurrent map + `-race` |
| `level_2/task_14` | or-channels (fan-in done) |
| `level_2/task_17` | telnet / TCP duplex |
| `level_2/task_18` | HTTP CRUD «Календарь» |

Дополнительно по вкусу: L1 `task_5` (таймер), L2 `task_9` (распаковка строки), L2 `task_16` (wget mirror).

## Что говорить

- Воркеры: «один канал задач, N читателей, закрытие канала = конец».
- Map: «Mutex вокруг map или sync.Map; без синка — data race».
- HTTP: handlers тонкие, валидация входа, коды 4xx/5xx осознанно.

После priority — добивай L2 утилиты (grep/cut/sort) если останется время.
