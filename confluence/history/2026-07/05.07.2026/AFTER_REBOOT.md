После ребута — поднять проект:

```bash
cd ~/event_horizon
make deploy
```

Проверить всё по порядку:

- Регистрация нового пользователя

- Логин и получение токена

- Смена никнейма

- Проверка баланса (лампочки/билетики)

- Проверка сохранения рекорда в Flappy

- WebSocket (лидерборд в реальном времени)

🧪 Команды для быстрой проверки:
```bash
# 1. Регистрация
curl -X POST http://localhost:8079/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test2@example.com","password":"secret123"}' | jq '.'

# 2. Логин
TOKEN=$(curl -s -X POST http://localhost:8079/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test2@example.com","password":"secret123"}' \
  | jq -r '.access_token')

# 3. Смена никнейма
curl -X POST http://localhost:8079/api/auth/update-nickname \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"nickname":"TestUser2"}' | jq '.'

# 4. Проверка баланса
curl -X GET http://localhost:8079/api/billing/balance/all \
  -H "Authorization: Bearer $TOKEN" | jq '.'

# 5. Лидерборд
curl -s "http://localhost:8079/api/leaderboard?game_id=hexagon&limit=10" | jq '.'
```

🎯 Фронтенд проверка
После ребута:

Запусти фронтенд:

```bash
cd ~/event_horizon/frontend
npm run dev
```

Открой http://localhost:5173

Зайди в профиль — проверь, что никнейм отображается корректно.

Смени ник через интерфейс (клик по нику) — проверь, что он обновился.

Проверь, что баланс отображается в профиле.

Сыграй в Flappy — проверь сохранение рекорда.