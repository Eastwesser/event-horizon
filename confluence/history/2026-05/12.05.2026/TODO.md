# TODO — 12 мая 2026

## Фронтенд (React + TypeScript)

### Утро: Настройка проекта
- [ ] Создать React приложение
  ```bash
  cd ~/event_horizon/frontend
  npm create vite@latest . -- --template react-ts
  npm install
Установить зависимости
bash
npm install axios          # HTTP клиент
npm install react-dnd      # Drag & drop
npm install react-dnd-html5-backend
npm install react-router-dom
npm install @tanstack/react-query  # Кеширование запросов
Настроить прокси для API (vite.config.ts)
typescript
export default defineConfig({
  server: {
    proxy: {
      '/api': 'http://localhost:8080',
      '/ws': {
        target: 'ws://localhost:8080',
        ws: true
      }
    }
  }
})
День: Компоненты

Auth компоненты
Форма регистрации
Форма логина
Хранение JWT в localStorage
Игровое поле (гексагоны)
Drag & drop реализация
Отрисовка гексагональной сетки
Блинчики с начинками (по манифесту!)
Leaderboard
WebSocket подключение
Отображение топа-10 в реальном времени
Billing UI
Отображение баланса (лампочки, билетики)
Кнопки "Купить подсказку", "Сбросить партию"
Вечер: Интеграция

Подключение к Gateway API
Тестирование сквозного сценария:
text
Регистрация → Логин → Игра → Отправка рекорда → 
Обновление топа через WebSocket → Начисление валюты
Структура фронтенда

text
frontend/
├── src/
│   ├── components/
│   │   ├── Auth/
│   │   │   ├── Login.tsx
│   │   │   └── Register.tsx
│   │   ├── Game/
│   │   │   ├── HexGrid.tsx      # Гексагональная сетка
│   │   │   ├── Tile.tsx         # Блинчик (drag-n-drop)
│   │   │   └── GameBoard.tsx
│   │   ├── Leaderboard/
│   │   │   └── Leaderboard.tsx  # WebSocket подключение
│   │   └── Billing/
│   │       └── Balance.tsx      # Лампочки/билетики
│   ├── hooks/
│   │   ├── useAuth.ts
│   │   ├── useWebSocket.ts
│   │   └── useGame.ts
│   ├── services/
│   │   ├── api.ts               # Axios инстанс
│   │   └── websocket.ts
│   ├── types/
│   │   └── index.ts             # TypeScript типы
│   ├── App.tsx
│   └── main.tsx
├── index.html
├── package.json
├── vite.config.ts
└── tsconfig.json
После фронтенда (13 мая+)

Notification сервис (push/email)
Analytics сервис (ClickHouse)
Prometheus + Grafana мониторинг
Payment сервис (Boosty)
Команды для быстрого старта фронтенда

bash
# Создание проекта
cd ~/event_horizon/frontend
npm create vite@latest . -- --template react-ts
npm install

# Установка зависимостей
npm install axios react-dnd react-dnd-html5-backend react-router-dom @tanstack/react-query

# Запуск фронтенда
npm run dev

# Сборка для продакшена
npm run build
План на 12 мая: Минимально жизнеспособный фронтенд

Регистрация/логин
Отображение гексагонального поля
Drag & drop для блинов
Отправка рекорда
WebSocket leaderboard
Отображение баланса
Критерий успеха к вечеру вторника: Пользователь может зарегистрироваться, сыграть одну партию, увидеть свой рекорд в топе и получить награду.