# Запуск программ в Linux (глазами Go)

## 1. execve — системный вызов, который заменяет процесс

- Не возвращается при успехе
- Заменяет память процесса на новый бинарник
- В Go: `syscall.Exec()` (редко), `os/exec` использует fork+exec

## 2. PATH

- `./program` — запуск из текущей папки
- `program` — поиск в `$PATH`
- В Go: `exec.LookPath()`

## 3. Права на выполнение (chmod +x)

- Без +x → permission denied
- Проверка в Go: `info.Mode() & 0111`

## 4. Shebang (#!/usr/bin/env python3)

- Ядро видит `#!` → запускает интерпретатор
- В Go: просто `exec.Command("./script.py")`, ядро само разбирается

## 5. Как ядро отличает ELF от скрипта

- Первые байты: `\x7fELF` → ELF
- `#!` → скрипт
- Остальное → Exec format error

## 6. ELF структура

- **Header**: магия, тип (ET_EXEC/ET_DYN), точка входа
- **Program headers**: как загружать в память (`readelf -l`)
- **Section headers**: для линковщика/отладки (`readelf -S`)
- **Секции**: .text (код), .rodata (константы), .data (инициализированные данные), .bss (zero values)

## 7. ET_EXEC vs ET_DYN (PIE)

- ET_EXEC: фиксированный адрес (уязвим для ASLR)
- ET_DYN: можно загрузить по любому адресу
- Go >=1.20: PIE по умолчанию

## 8. Полезные команды

```bash
readelf -h /bin/ls
readelf -l myapp
objdump -d myapp | head -50
hexdump -C myapp | head -1  # проверить магию
strace -e execve /bin/ls
```