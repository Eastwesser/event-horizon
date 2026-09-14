# Урок 14 — 11 мая 2026

Короткие concurrency/string сниппеты из конспекта.

## Задачи

### 1. Fan-out из слайса
Значения слайса → буферизованный канал (или N воркеров). Фраза: «сначала наполняем jobs / закрываем; читатели range; WaitGroup если fan-out в workers».

### 2. `strStr` (подстрока)
Есть ли `needle` в `haystack`. На собесе: naive O(n·m) ок для скрининга; упомяни KMP/Rabin-Karp как «знаю, что есть лучше». В Go можно честно сказать про `strings.Contains`, но здесь — руками.

## Как запускать

```bash
export GOWORK=off
cd code/01_fan_out_slice && go run main.go
```
