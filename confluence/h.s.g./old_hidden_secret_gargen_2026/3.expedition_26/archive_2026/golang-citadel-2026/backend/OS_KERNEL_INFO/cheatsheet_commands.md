# Шпаргалка: системные команды для Go-разработчика

## Процессы
```bash
# PID своего процесса
echo $$

# Дерево процессов
pstree -p <PID>

# Потоки процесса
ls /proc/<PID>/task | wc -l
ps -eLf | grep <PID>
```

## Память
```bash
# Виртуальная и физическая память процесса
cat /proc/<PID>/status | grep -E "Vm|Rss"

# Карта памяти (все VMA)
cat /proc/<PID>/maps

# Суммарно по типам
cat /proc/<PID>/smaps | grep -E "Pss|Rss"
```

## Системные вызовы
```bash
# Следить за execve
strace -e execve ./program

# Следить за clone (создание потоков)
strace -f -e clone ./program

# Следить за mmap
strace -e mmap,munmap ./program

# Всё подряд для запущенного процесса
strace -f -p <PID>
```

## ELF-анализ
```bash
# Заголовок ELF
readelf -h /bin/ls

# Program headers (сегменты)
readelf -l /bin/ls

# Section headers (секции)
readelf -S /bin/ls

# Дизассемблер
objdump -d /bin/ls | head -100

# Размер секций
size -A /bin/ls
```

## Go-specific
```bash
# Информация о GC и памяти во время работы
GODEBUG=gctrace=1 ./myapp

# Сборка с отладочными символами
go build -gcflags="-N -l" -o myapp main.go

# Профилирование памяти
go test -bench . -memprofile mem.out
go tool pprof -http=:8080 mem.out
```

## Инструменты ядра
```bash
# Следить за page fault'ами
perf stat -e page-faults ./myapp

# Следить за системными вызовами в реальном времени
perf trace ./myapp
```
