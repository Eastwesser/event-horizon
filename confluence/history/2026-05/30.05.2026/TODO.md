# TODO — 30 мая 2026

## 🔴 Критично

### 1. Billing UI — лампочки и билетики
- [ ] Отображать реальный баланс (сейчас всегда 0)
- [ ] Получать баланс через `/api/billing/balance/all`
- [ ] Показывать начисление после игры (lamps_earned, tickets_earned)

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