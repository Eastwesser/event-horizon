🔧 ПРОМПТ ДЛЯ CURSOR: ИСПРАВЛЕНИЕ GO WORKSPACE И СБОРКИ
markdown
# Agent: Event-Horizon Build System Fixer

Ты — **инженер по сборке и инфраструктуре**. Твоя задача — починить `make deploy` и настроить **стабильную систему сборки** для всех сервисов.

---

## 🚨 ТЕКУЩАЯ ПРОБЛЕМА

При попытке собрать проект через `make deploy` возникает ошибка:
stat /home/denismatveev/event_horizon/services/nats-hub/main.go: directory not found

text

А также конфликт версий Go:
module ../mcp listed in go.work file requires go >= 1.26.5, but go.work lists go 1.25.7

text

**Причина:** В проекте есть **go.work** файл, который ссылается на несуществующую папку `../mcp` и требует Go 1.26.5, хотя используется Go 1.25.7.

---

## 🎯 ЗАДАЧА

Починить систему сборки так, чтобы:

1. `make deploy` работал без ошибок
2. Все сервисы собирались (nats-hub, auth, game, billing, и т.д.)
3. go.work был корректным и не ссылался на несуществующие модули
4. Версия Go была единой для всего проекта

---

## 🔍 АУДИТ: ЧТО НУЖНО ПРОВЕРИТЬ

### 1. Проверить go.work

```bash
# Посмотреть текущий go.work
cat go.work

# Проверить, какие модули там перечислены
go work edit -json
Ожидаемый результат: go.work должен содержать только реально существующие папки:

text
go 1.25.7

use (
    ./services/auth
    ./services/game
    ./services/billing
    ./services/leaderboard
    ./services/profile
    ./services/shop
    ./services/inventory
    ./services/gateway
    ./services/balancer
    ./services/nats-hub
    ./services/payment
    ./services/authors
    ./services/analytics
    ./pkg/redisclient
    ./pkg/errors
)
2. Проверить go.mod каждого сервиса
bash
# Проверить версию Go в каждом сервисе
grep -r "go 1." services/*/go.mod
Ожидаемый результат: во всех сервисах должна быть одна и та же версия (например, go 1.25.7).

3. Проверить main.go в nats-hub
bash
# Проверить, существует ли файл
ls -la services/nats-hub/main.go

# Если нет — создать
Проблема: make ищет main.go в services/nats-hub/, а его там нет (возможно, он был перенесён или удалён).

Решение: либо создать main.go, либо обновить Makefile, чтобы он искал правильный путь.

🔧 РЕШЕНИЕ (ПОШАГОВО)
Шаг 1: Обновить go.work
bash
# Удалить старый go.work
rm go.work
go.work.sum

# Создать новый с правильными путями
go work init \
    ./services/auth \
    ./services/game \
    ./services/billing \
    ./services/leaderboard \
    ./services/profile \
    ./services/shop \
    ./services/inventory \
    ./services/gateway \
    ./services/balancer \
    ./services/nats-hub \
    ./services/payment \
    ./services/authors \
    ./services/analytics \
    ./pkg/redisclient
ИЛИ вручную отредактировать go.work:

text
go 1.25.7

use (
    ./services/auth
    ./services/game
    ./services/billing
    ./services/leaderboard
    ./services/profile
    ./services/shop
    ./services/inventory
    ./services/gateway
    ./services/balancer
    ./services/nats-hub
    ./services/payment
    ./services/authors
    ./services/analytics
    ./pkg/redisclient
)
Шаг 2: Обновить go.mod каждого сервиса
Убедись, что в каждом go.mod указана одинаковая версия Go:

go
module github.com/Eastwesser/event-horizon/services/auth

go 1.25.7  // ← должно быть одинаково везде
Шаг 3: Проверить и создать main.go для nats-hub
Если services/nats-hub/main.go не существует — создай его:

go
// services/nats-hub/main.go
package main

import (
    "log"
    "os"
    "time"

    "github.com/nats-io/nats.go"
)

func main() {
    natsURL := os.Getenv("NATS_URL")
    if natsURL == "" {
        natsURL = "nats://localhost:4222"
    }

    nc, err := nats.Connect(natsURL)
    if err != nil {
        log.Fatalf("Failed to connect to NATS: %v", err)
    }
    defer nc.Close()

    js, err := nc.JetStream()
    if err != nil {
        log.Fatalf("Failed to create JetStream context: %v", err)
    }

    // Создаём Stream для событий
    _, err = js.AddStream(&nats.StreamConfig{
        Name:     "EVENTS",
        Subjects: []string{
            "event.>",
            "score.updated",
            "user.registered",
            "shop.purchased",
            "payment.completed",
        },
        Storage:  nats.FileStorage,
        MaxAge:   7 * 24 * time.Hour,
    })
    if err != nil {
        log.Printf("Stream might already exist: %v", err)
    } else {
        log.Println("✅ Stream EVENTS created")
    }

    // Держим сервис запущенным
    select {}
}
Шаг 4: Обновить Makefile
Проверь, что в Makefile правильные пути для сборки:

makefile
# Makefile
build-nats-hub:
	@echo "🔨 Building nats-hub..."
	cd services/nats-hub && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o nats-hub .
	@echo "✅ nats-hub built"
Шаг 5: Обновить go.mod для nats-hub
Убедись, что у nats-hub есть правильный go.mod:

bash
cd services/nats-hub
go mod init github.com/Eastwesser/event-horizon/services/nats-hub
go mod tidy
Шаг 6: Синхронизировать зависимости
bash
# Обновить все зависимости
go work sync

# Проверить, что всё собирается
go build ./...
✅ ФИНАЛЬНЫЙ ЧЕК-ЛИСТ
Перед тем как запустить make deploy, проверь:

□ go.work существует и содержит все сервисы
□ go.work НЕ содержит ссылок на несуществующие папки (например, ../mcp)
□ Во всех go.mod одинаковая версия Go (1.25.7)
□ services/nats-hub/main.go существует
□ services/nats-hub/go.mod существует
□ go work sync выполнен без ошибок
□ go build ./... работает без ошибок
📋 КОМАНДЫ ДЛЯ ПРОВЕРКИ
bash
# 1. Проверить go.work
cat go.work

# 2. Проверить версии Go
grep -r "go 1." services/*/go.mod | sort -u

# 3. Проверить nats-hub
ls -la services/nats-hub/

# 4. Проверить сборку
go build ./...

# 5. Запустить деплой
make deploy
🚫 ЧЕГО НЕ ДЕЛАТЬ
❌ Не удалять go.work без создания нового

❌ Не использовать разные версии Go в разных сервисах

❌ Не добавлять в go.work папки, которых нет в проекте

❌ Не игнорировать ошибки go work sync

📋 СТИЛЬ ОТВЕТА
После выполнения шагов, пришли:

bash
# 1. Содержимое go.work
cat go.work

# 2. Результат проверки версий
grep -r "go 1." services/*/go.mod | sort -u

# 3. Результат сборки
go build ./...
make deploy
Время пошло. Начинай с go.work. 🚀

text

---

## 📌 ДОПОЛНИТЕЛЬНЫЙ СОВЕТ

Если после всех исправлений всё ещё будут проблемы — проверь:

1. **Go version**: `go version` — должна быть 1.25.7 или выше
2. **GOPATH**: `echo $GOPATH` — не должен указывать на папку проекта
3. **GO111MODULE**: `echo $GO111MODULE` — должно быть `on` или не задано

**Альтернативный вариант:** если `go.work` мешает, можно временно отключить его:

```bash
# Отключить go.work
go env -w GOWORK=off

# Собрать проект
make deploy

# Включить обратно (если нужно)
go env -w GOWORK=on
Но лучше всё-таки починить go.work, чтобы проект собирался стандартным способом.