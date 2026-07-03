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