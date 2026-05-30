# TODO — 30 мая 2026

## 🔴 Критично

### 2. Выбор уровня сложности
- [ ] Компонент LevelSelector (1-20)
- [ ] Множитель очков: `score * (level / 10)` или `score * sqrt(level)`
- [ ] Передавать уровень в /api/game/submit

### 3. Анимация пуф-эффекта
- [ ] CSS-анимация при очистке стопки (≥10 блинов)
- [ ] Длительность 0.5-1 секунда

## 🟡 Важно

### 4. Goose миграции для других сервисов
- [ ] Добавить миграции для Game, Billing, Leaderboard
- [ ] Автоматический запуск при `make all`

### 5. Нагрузочное тестирование
- [ ] Go producer + consumer для NATS
- [ ] k6 + NATS extension
- [ ] Мониторинг очереди JetStream

## 🟢 Приятно

### 6. Улучшения UI/UX
- [ ] Spinner загрузки в лидерборде
- [ ] Обработка ошибок соединения
- [ ] Подсветка возможных ходов

### 7. NATS кластер из 3 нод

## Команды для запуска

```bash
# Полный старт
cd ~/event_horizon
make all

# Фронтенд
cd frontend && npm run dev

# Проверить логи
tail -f /tmp/game.log
tail -f /tmp/leaderboard.log
tail -f /tmp/gateway.log

# NATS подписка
nats sub "score.updated" --server localhost:4222

Осталось в техдолге (записываю):

Персистентное хранение статистики на бэкенде — чтобы профиль не сбрасывался при очистке localStorage
Выбор уровня сложности (1-20) перед игрой
Анимация пуф-эффекта при очистке стопки
Блинопекарня (магазин за лампочки)
CQRS для leaderboard (улучшение)


Добавляю в техдолг:

markdown
### Лидерборд — отображение кастомного ника

**Проблема:** Сейчас в лидерборде отображается `user_email.split('@')[0]` (например, "vinland"). Если пользователь сменил ник в профиле на "V1nl@nD", в топе всё равно видно старый email-based ник.

**Требование:**
- При отправке рекорда передавать на бэкенд текущий `nickname` из localStorage
- Бэкенд должен сохранять последний использованный ник для каждого user_id
- В лидерборде отображать кастомный ник, если он есть, иначе fallback на email

**Что нужно изменить:**

1. **Фронтенд** — в `submitScore` добавить поле `nickname`:
```typescript
const nickname = localStorage.getItem('nickname') || email.split('@')[0];

await api.post('/game/submit', {
  user_id: userId,
  game_id: 'hexagon',
  level: level,
  score: currentScore,
  user_email: userEmail,
  nickname: nickname,  // 👈 добавить
  seed: 'game_seed_' + Date.now(),
  moves: [],
});
Бэкенд (Game service) — принимать nickname в запросе и пробрасывать в NATS событие
Бэкенд (Leaderboard service) — сохранять nickname в Redis вместе с счётом:
text
ZADD leaderboard:hexagon <score> <user_id>
HSET user:profile:<user_id> nickname <nickname> email <email>
API лидерборда — возвращать nickname из профиля, если есть
Приоритет: 🟡 Средний (после критичных багов)
Оценка: ~2-3 часа (бэкенд + фронтенд)

text

---

## Итоговый техдолг по профилю и лидерборду:

| # | Задача | Приоритет | Статус |
|---|--------|-----------|--------|
| 1 | Профиль на фронтенде (localStorage) | 🟢 Сделано | ✅ |
| 2 | Персистентное хранение статистики на бэкенде | 🟡 Средний | 📝 |
| 3 | Кастомный ник в лидерборде | 🟡 Средний | 📝 |
| 4 | Синхронизация ника между устройствами | 🟢 Низкий | 📝 |
