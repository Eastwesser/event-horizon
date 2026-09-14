# Урок 02 — 18 марта 2026

Code-review кэша под highload (20% write / 80% read), баг type-assert в LRU-обёртке, range+переприсвоение слайса, singleflight к платному API, IncrementUniqueBy, Once на каналах.

## Задачи

### 1. Concurrent cache (RWMutex)
Локальный `sync.Mutex` в функции — мёртвый; нужен shared mutex/структура. Write под `Lock`, read под `RLock`. Double-check locking в GetOrCreate; sync.Map / шарды — уровень 20.

### 2. Cache type assert
В кэш кладут value `Warehouse`, а достают `*Warehouse` → miss → БД не разгружается. Клади и читай один и тот же тип (обычно указатель).

### 3. `range` + переприсвоение слайса
`range` фиксирует длину и ходит по исходному массиву; `lst = newSlice` не меняет итерацию. Вывод: a b c d.

### 4. Singleflight
Одинаковые in-flight запросы к платному API схлопываются: первый ходит наружу, остальные ждут `done`. Mutex + map + channel.

### 5. IncrementUniqueBy
По указателям: уникальный адрес инкрементируем один раз; считаем updated/duplicates; nil пропускаем.

### 6. Once на каналах
Буферизованный chan token: кто забрал — выполняет f. Без `sync.Once`.

## Как запускать

```bash
export GOWORK=off
cd code/01_concurrent_cache && go run main.go
```
