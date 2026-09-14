# mmap и управление памятью в Go

## mmap — системный вызов для управления виртуальной памятью
- Резервирует **виртуальную** память (мгновенно)
- Физическая RAM выделяется **лениво** (при первом касании страницы)
- В Go: `syscall.Mmap()`

## Ключевые понятия
- **VmSize** (`/proc/pid/status`) — виртуальная память (все резервации)
- **VmRSS** — физическая RAM (реально используемая)
- **Page fault** — исключение, когда страница есть виртуально, но нет физически

## Эксперимент: выделить 10GB виртуальной
```go
data, _ := syscall.Mmap(-1, 0, 10<<30, 
    syscall.PROT_READ|syscall.PROT_WRITE, 
    syscall.MAP_PRIVATE|syscall.MAP_ANONYMOUS)
// VmSize +10GB, VmRSS ~0

// Touch pages:
for i := 0; i < 10<<30; i += 4096 {
    data[i] = 0xAA  // page fault → физическая RAM
}
// VmRSS ≈ VmSize
Как Go использует mmap
Heap (арены по 64MB) — выделяются через mmap

Стеки горутин — лежат в heap (не в стеке треда!)

Memory-mapped файлы — через os.Mmap()

Структура heap в Go (mheap)
Arenas — 64MB куски виртуальной памяти

Spans — 8KB куски внутри арены

Объекты: tiny (<16B), small (16B-32KB), large (>32KB)

Полезные команды
bash
# Смотреть все VMA процесса
cat /proc/<pid>/maps

# Статистика памяти
cat /proc/<pid>/status | grep -E "Vm|Rss"

# Следить за mmap вызовами
strace -e mmap,munmap go run main.go
Выводы для Go-разработчика
HeapSys может быть большим, а HeapAlloc маленьким — память виртуально зарезервирована, но физически не выделена

GC не всегда возвращает память ОС (munmap) — держит в mheap для будущих аллокаций

Стек горутины может расти (до 1GB) через page fault

make([]byte, 1e9) не выделяет 1GB физической RAM сразу