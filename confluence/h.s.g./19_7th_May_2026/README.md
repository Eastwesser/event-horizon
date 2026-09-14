# Урок 19 — map internals + Denis Sher ката

## Тема
Заметки про новые Go maps (SIMD / top-hash / tombstone) + практические задачи Ката Босса. Исходники также в `denis_sher_tasks/denis_sher_t/` — здесь **чистые runnable копии** в `code/`.

## Задачи (`code/`)

| Pack | Что |
|------|-----|
| `01_append_slice_puzzle` | `y:=append(x,3); z:=append(x,4)` — общий backing |
| `02_semaphore` | ≤5 одновременных «скачиваний» из 100 |
| `03_mutateSlice` | mutate через header vs reallocation |
| `04_mergeChans` | fan-in N→1 |
| `05_sliceFilter` | filter in-place без `make` |
| `06_slicePointers` | `*Person` vs переприсвоение указателя |
| `07_whatsWrong` | loop var + timeout + Body.Close |

Оригинал: `denis_sher_tasks/denis_sher_t/{1..6}/`.

## Как решать / что говорить
- **Slice:** «header {ptr,len,cap} копируется; append при `len==cap` даёт новый массив».
- **Semaphore:** `chan struct{}` ёмкости N — Acquire/Release.
- **Merge:** каждый вход — горутина + `WaitGroup` + один `close(out)`.
- **Maps (теория):** top-hash в meta-байте, tombstone при delete, реже evacuation в SwissTable.

## Собес-советы
- Нарисуй header слайса на доске перед кодом.
- Для `whatsWrong` перечисли 3 бага вслух до правок.
- Видео про maps: см. ссылку в `19_full_task.md`.

## Как запускать

Путь содержит `h.s.g.` (точки) — `go` может подхватить `go.work` монорепо и сломаться. Всегда:

```bash
cd code/NN_name
export GOWORK=off
go run main.go
```

