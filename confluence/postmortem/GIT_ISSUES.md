# ПЛАН А: ЧИНИМ ГОЛОВУ

cd /home/denismatveev/event_horizon

## 1. Удаляем пустые объекты (как в прошлый раз)
find .git/objects/ -type f -empty -delete
find .git/objects/ -size 0 -delete

## 2. Восстанавливаем HEAD из бэкапа (если есть)
cp .git.backup/HEAD .git/HEAD 2>/dev/null
echo "ref: refs/heads/main" > .git/HEAD

## 3. Восстанавливаем ссылки на ветки из бэкапа
cp .git.backup/refs/heads/main .git/refs/heads/main 2>/dev/null
cp .git.backup/refs/remotes/origin/main .git/refs/remotes/origin/main 2>/dev/null

## 4. Пытаемся починить
git fsck --full --no-dangling

## 5. Сборщик мусора (если получится)
git gc --prune=now --force

### =====================================

# ПЛАН Б: Когда все совсем плохо:

cd /home/denismatveev

## 1. Сохраняем все наши изменения (фронтенд, бэкенд, документы)
cp -r event_horizon event_horizon_save

## 2. Свежий клон (глубокий, чтобы не тащить всю историю)
git clone --depth 1 https://github.com/Eastwesser/event-horizon.git event_horizon_new

## 3. Копируем сохранённые файлы обратно
cp -r event_horizon_save/frontend/src/components/Games/ event_horizon_new/frontend/src/components/
cp -r event_horizon_save/services/gateway/internal/ratelimit/ event_horizon_new/services/gateway/internal/
cp -r event_horizon_save/services/gateway/internal/middleware/ event_horizon_new/services/gateway/internal/
cp event_horizon_save/confluence/* event_horizon_new/confluence/ -r

## 4. Удаляем старый и переключаемся на новый
rm -rf event_horizon
mv event_horizon_new event_horizon

#  План C =====

## 📅 Дата
2026-07-03

## 🐛 Проблема
Git-репозиторий повреждён: объект `44715dc4c8ef22e4883aaeb96c501bb48417d2dd` (HEAD) стал пустым.
- `git status` выдавал `fatal: bad object HEAD`
- `git fsck` показывал `invalid sha1 pointer`
- `git gc` падал с `fatal: bad tree object`

## 🔍 Причина
Скорее всего, сбой при записи объекта в `.git/objects/` (аварийное завершение, проблемы с диском, прерванный `git push`).

---

## 🛠️ Решение (ПЛАН Б — Полный перезаезд)

### 1. Удаление пустых объектов и индекса

```bash
find .git/objects/ -type f -empty -delete
rm -f .git/index
echo "ref: refs/heads/main" > .git/HEAD
git symbolic-ref HEAD refs/heads/main
git reset
```

2. Создание нового коммита со всеми изменениями

```bash
git add .
git commit -m "🔥 v1.0.2: Полный мониторинг и метрики (восстановление)"
```

3. Разрешение конфликтов при пуше

```bash
git pull --no-ff --no-commit origin main
```

Возникли конфликты в: auth, balancer, billing, game, leaderboard, store/README.md

Решение: оставить локальные версии (т.к. они свежее):

```bash
git checkout --ours services/auth/cmd/main.go
git checkout --ours services/balancer/cmd/main.go
git checkout --ours services/billing/cmd/main.go
git checkout --ours services/game/cmd/main.go
git checkout --ours services/leaderboard/cmd/main.go
git checkout --ours services/store/README.md
git add .
git commit -m "Merge: resolve conflicts (keep local versions)"
```

4. Финальный пуш

```bash
git push origin main
```

---

🛡️ Предотвращение в будущем
bash
# Добавить в ~/.bashrc
alias git-safe='cp -r .git .git.backup_$(date +%Y%m%d_%H%M%S) && git'