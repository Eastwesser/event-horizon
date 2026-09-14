# Урок 22 — training checklist (study guide)

## Тема
Чек-лист повторения перед собесом: TCP/UDP, TLS, sync, каналы, ошибки, БД. Код здесь — **маленькие sync-демо**, не сеть.

Сырьё: `22_training_checklist.md`.

## Задачи (`code/`)

| Pack | Что потренировать |
|------|-------------------|
| `01_waitgroup` | Add/Done/Wait |
| `02_channel_close` | close писателем, range, ok=false |
| `03_mutex_once` | Mutex / RWMutex / Once |

## Study guide (проговорить вслух)
1. **TCP vs UDP** — handshake, надёжность, порядок; логи с батчингом vs стриминг.
2. **TLS** — ClientHello → ServerHello+cert → key exchange → symmetric.
3. **Каналы** — buffered/unbuffered, кто закрывает, select+default.
4. **Ошибки** — свой тип с `Error()`, `errors.Is/As`, `%w`.
5. **Порты** — 22/80/443/5432/6379/8080/9090.

## Собес-советы
- GoLand: отключи автодополнение на лайвкодинге.
- Перед звонком — 15 минут прогнать этот чек-лист + 3 демо ниже.

## Как запускать

Путь содержит `h.s.g.` (точки) — `go` может подхватить `go.work` монорепо и сломаться. Всегда:

```bash
cd code/NN_name
export GOWORK=off
go run main.go
```

