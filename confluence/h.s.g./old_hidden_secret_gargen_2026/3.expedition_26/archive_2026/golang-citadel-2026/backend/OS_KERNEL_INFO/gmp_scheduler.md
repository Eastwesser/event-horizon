# GMP: горутины, треды, процессоры

## Модель
- **G** (Goroutine) — легковесный поток (2KB стек)
- **M** (Machine) — OS thread (создаётся через `clone`)
- **P** (Processor) — логический процессор, планировщик (количество = GOMAXPROCS)

## Создание M (OS thread) — syscall clone
```c
clone(CLONE_VM | CLONE_FS | CLONE_FILES | CLONE_SIGHAND | CLONE_THREAD, ...)
```
- Общая память с родителем
- Общие файловые дескрипторы
- Один PID для всех нитей (ps -eLf)

## Почему горутины легче тредов
```text
                Горутина	            Тред
Стек	        2KB (растёт)	            8MB фиксированно
Создание	микросекунды (malloc)	    миллисекунды (clone)
10k штук	20MB RAM	            80GB RAM
```

## Когда создаётся новый M

- Блокирующий syscall (чтение с диска)
- CGO вызов
- runtime.LockOSThread()
- Нехватка M при наличии P

## Типы системных вызовов из горутины

- Быстрые (getpid) — прямой вызов, M не блокируется
- Сетевые (read от сокета) — неблокирующий режим + epoll, G паркуется
- Блокирующие (read от файла) — M уходит в блокировку, P получает другой M

## Ключевые функции рантайма (исходники Go)

- runtime/proc.go — планировщик
- runtime/os_linux.go — обёртки clone
- runtime/stack.go — управление стеками G

Команды для наблюдения
```bash
# Потоки процесса
ls /proc/<PID>/task | wc -l
ps -eLf | grep main
```

# Системные вызовы
strace -f -e clone,epoll_wait ./myapp

Пример: принудительное создание M
```go
// Блокирующий read из файла вызовет создание нового M
f, _ := os.Open("/dev/random")
go func() {
    buf := make([]byte, 1)
    f.Read(buf)  // новый M!
}()
```
