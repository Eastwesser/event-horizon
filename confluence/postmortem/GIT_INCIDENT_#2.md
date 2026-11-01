# Git: инциденты и восстановление

Этот документ собирает постмортемы по повреждениям Git-репозитория и рабочие рецепты восстановления. Каждый инцидент — отдельный раздел. В конце — общие рекомендации и «шпаргалка» на случай паники.

---

## 📅 Инцидент 2026-11-02: пустой объект `cf9b4923...`, сломанный HEAD

### 🐛 Симптомы
- `git status` / `git add` / `git commit` падали с:
error: object file .git/objects/cf/9b49231049f070037d7ca06026f6f5ac9ad7a0 is empty
fatal: loose object cf9b4923... is corrupt

text
- Позже: `fatal: не удалось разобрать HEAD`, `fatal: bad object HEAD`.
- `git reflog` не читался (`fatal: bad object HEAD`).
- `git fsck --full` показал ещё несколько пустых объектов:
`145af8...`, `22435b...`, `26c70f...`, `92b5ba...`, `97d6e3...`, `fb215a...`,
а также `invalid sha1 pointer` для `refs/heads/main`, `refs/remotes/origin/main`, `HEAD`.

### 🔍 Причина
Сбой записи объектов в `.git/objects/` (аварийное завершение, проблемы с диском, прерванный `push`/`gc`). В результате несколько loose-объектов оказались пустыми (0 байт), а `main` и `origin/main` указывали на несуществующий `cf9b4923...`. Это блокировало чтение `HEAD` и любые операции.

### 🛠️ Что делали (по шагам)

1. **Бэкап `.git`** (обязательно перед любой хирургией):
 ```bash
 cp -r .git .git-backup-$(date +%Y%m%d-%H%M%S)
В нашем случае бэкап не содержал нужных объектов, но это страховка.

Нашли все пустые объекты:

bash
find .git/objects -type f -empty
Удалили пустые объекты (они всё равно нерабочие):

bash
find .git/objects -type f -empty -delete
Распаковали pack-файлы — это вернуло тысячи объектов из упакованного хранилища:

bash
for pack in .git/objects/pack/*.pack; do
    git unpack-objects -r < "$pack"
done
Флаг -r (recover) позволяет продолжить распаковку даже при повреждениях в pack-файле.

Удалили сломанные remote-ссылки (они восстановятся при fetch):

bash
git update-ref -d refs/remotes/origin/main
git update-ref -d refs/remotes/origin/HEAD
Нашли «потерянные» коммиты через fsck:

bash
git fsck --lost-found --no-reflogs 2>/dev/null | grep "dangling commit"
Получили список из 5 dangling-коммитов. Посмотрели их:

bash
for c in <хеши>; do
    git log -1 --format="%h %ad %s" --date=iso "$c"
done
Самый свежий (3a542cf, 2026-10-27, «docs: updated redis for game - anticheat feature for v.1.1.4 and +») оказался вершиной истории main.

Переставили main на здоровый коммит и починили HEAD:

bash
git update-ref refs/heads/main 3a542cfa2705bc4301caf029a39af3bd75fe044b
git symbolic-ref HEAD refs/heads/main
После этого git status ожил, изменения (staged/unstaged) оказались на месте.

Закоммитили изменения (два коммита: .gitignore + docs).

Синхронизировались с сервером:

bash
git pull --rebase origin main
git push
push показал cf9b492..00ddf6b main -> main — то есть удалённая main действительно указывала на битый cf9b4923..., теперь перезаписана восстановленной историей.

✅ Результат
История коммитов сохранена (древо не сносили).

main и origin/main указывают на здоровые коммиты.

Рабочие изменения закоммичены и запушены.

Репозиторий синхронизирован.

🧠 Выводы
Пустой объект в .git/objects/ — это не приговор: часто достаточно удалить пустышки и распаковать pack-файлы.

git fsck --lost-found — главный инструмент поиска «потерянных» коммитов.

git update-ref + git symbolic-ref — аккуратный способ починить main и HEAD без пересоздания репозитория.

git pull --rebase решает проблему «remote contains work you do not have» мягко, без --force.

📅 Инцидент 2026-07-03: пустой объект 44715dc4... (HEAD)
🐛 Симптомы
git status: fatal: bad object HEAD

git fsck: invalid sha1 pointer

git gc: fatal: bad tree object

🔍 Причина
Сбой записи объекта в .git/objects/ (аварийное завершение, проблемы с диском, прерванный git push).

🛠️ Решение (План Б — «полный перезаезд»)
Удалить пустые объекты и индекс, пересоздать HEAD:

bash
find .git/objects/ -type f -empty -delete
rm -f .git/index
echo "ref: refs/heads/main" > .git/HEAD
git symbolic-ref HEAD refs/heads/main
git reset
Создать новый коммит со всеми изменениями:

bash
git add .
git commit -m "🔥 v1.0.2: Полный мониторинг и метрики (восстановление)"
Разрешить конфликты при пуше:

bash
git pull --no-ff --no-commit origin main
# при конфликтах — оставить локальные версии:
git checkout --ours <файлы>
git add .
git commit -m "Merge: resolve conflicts (keep local versions)"
Финал:

bash
git push origin main
🧠 Выводы
Если HEAD указывает на пустой объект, а reflog не читается — можно пересоздать HEAD и индекс, а затем закоммитить текущее состояние рабочей директории.

«Полный перезаезд» (План Б) — крайняя мера, когда история не критична или её не удаётся восстановить. В инциденте 2026-11-02 он не понадобился: историю спасли через fsck + update-ref.

🧰 Общая шпаргалка: что делать, если Git «сломался»
0. Не паниковать и сделать бэкап
bash
cp -r .git .git-backup-$(date +%Y%m%d-%H%M%S)
Не удаляй бэкап, пока не убедишься, что всё стабильно.

1. Диагностика
bash
# Какие объекты пустые?
find .git/objects -type f -empty

# Полная проверка целостности
git fsck --full

# Поиск «потерянных» коммитов
git fsck --lost-found --no-reflogs 2>/dev/null | grep "dangling commit"

# Что с HEAD и ветками?
cat .git/HEAD
cat .git/refs/heads/main
git rev-parse HEAD
2. Лечение (от мягкого к жёсткому)
Удалить пустые объекты и распаковать pack-файлы:

bash
find .git/objects -type f -empty -delete
for pack in .git/objects/pack/*.pack; do
    git unpack-objects -r < "$pack"
done
Починить ссылки:

bash
# удалить сломанные remote-ссылки
git update-ref -d refs/remotes/origin/main
git update-ref -d refs/remotes/origin/HEAD

# переставить main на здоровый коммит (из fsck --lost-found)
git update-ref refs/heads/main <хороший-хеш>
git symbolic-ref HEAD refs/heads/main
Синхронизация с сервером:

bash
git fetch origin
git pull --rebase origin main
git push
Крайняя мера — «полный перезаезд» (План Б): сохранить рабочую директорию, сделать свежий clone, вернуть файлы, закоммитить. Использовать, только если история не нужна или не восстанавливается.

3. Полезные команды на каждый день
bash
# История без пагинатора
git --no-pager log --oneline -10
git -c core.pager=cat log --oneline -10

# Проверить, что репозиторий жив
git rev-parse HEAD
git branch --show-current
git fetch origin && git status

# Безопасный пуш после rebase
git push --force-with-lease origin main
4. Профилактика
Не прерывать git gc, git push, git commit (особенно на слабом диске).

Периодически проверять git fsck --full на важных репозиториях.

Алиас для безопасных операций:

bash
# ~/.bashrc
alias git-safe='cp -r .git .git.backup_$(date +%Y%m%d_%H%M%S) && git'
Держать .git-backup-*/ в .gitignore, чтобы бэкапы не попадали в коммиты.

📌 Итог
Инцидент 2026-11-02 показал: даже при нескольких пустых объектах и сломанном HEAD историю можно спасти через fsck --lost-found + update-ref, без сноса проекта.

«План Б» (полный перезаезд) остаётся на крайний случай.

Главное — бэкап .git, диагностика через fsck и аккуратная починка ссылок.