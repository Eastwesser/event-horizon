RUNBOOK_INTERNET_v1.1.3.md — как выставить Event Horizon в интернет (Quick Tunnel)
Версия: v1.1.3, «по-уебищному, но работает». Когда будут бабки — перейдём на Named Tunnel + домен.

Шаг 0. Предусловия
Виртуалка (Arch Linux) запущена.

Docker работает.

cloudflared установлен (sudo pacman -S cloudflared).

В frontend/vite.config.ts нет allowedHosts с хардкодом (или есть, но мы всё равно передаём переменную).

Шаг 1. Поднять бэкенд
bash
cd ~/event_horizon
make deploy
Дождаться:

text
✅ Thin stack ready
   🎯 Gateway: http://localhost:8079
Проверить:

bash
curl -fsS http://localhost:8079/health
# {"service":"gateway","status":"ok"}
Шаг 2. Поднять фронтенд (Vite) — пока без переменной
bash
cd ~/event_horizon/frontend
npm run dev
Дождаться:

text
  ➜  Local:   http://localhost:5173/
Терминал не закрывать. Vite работает.

Шаг 3. Поднять Quick Tunnel на Vite (в новом терминале)
bash
# Убить старый туннель, если висит
pkill -f 'cloudflared tunnel --url http://localhost:5173' || true

# Запустить новый
nohup cloudflared tunnel --url http://localhost:5173 > ~/eh-frontend-tunnel.log 2>&1 &

# Подождать
sleep 3

# Достать ссылку
grep -o 'https://[^ ]*\.trycloudflare\.com' ~/eh-frontend-tunnel.log | tail -n 1
Получишь ссылку вида:

text
https://desktop-euro-bathroom-oriented.trycloudflare.com
Скопируй её.

Шаг 4. Перезапустить Vite с этой ссылкой
В терминале, где работает Vite, нажми Ctrl+C. Потом:

bash
__VITE_ADDITIONAL_SERVER_ALLOWED_HOSTS="desktop-euro-bathroom-oriented.trycloudflare.com" npm run dev
(подставь свою ссылку из Шага 3)

Дождаться:

text
  ➜  Local:   http://localhost:5173/
Терминал не закрывать.

Шаг 5. Проверить с телефона (не с виртуалки)
Открыть в браузере телефона:

text
https://desktop-euro-bathroom-oriented.trycloudflare.com
Должен открыться фронтенд. Если открылся — всё работает. Можно скинуть жене.

Шаг 6. Выключить (когда закончил)
В терминале с Vite: Ctrl+C.

В терминале с туннелем:

bash
pkill -f 'cloudflared tunnel --url http://localhost:5173'
Бэкенд можно оставить или make down.

Ссылка мгновенно перестанет работать.

⚠️ Что помнить
Ссылка меняется каждый раз, когда ты запускаешь cloudflared заново. Завтра будет другая. Это нормально для Quick Tunnel.

Порядок важен: сначала Vite, потом туннель, потом Vite перезапустить с новой ссылкой. Если перезапустить Vite до туннеля — он не будет знать ссылку.

Vite и cloudflared должны работать одновременно. Закрыл Vite — ссылка умерла (даже если cloudflared жив). Закрыл cloudflared — ссылка умерла (даже если Vite жив).

Precheck cloudflared может ругаться на region2 — это шум. Если ссылка появилась и открывается — всё ок.

Ничего не покупаем, ничего не настраиваем. Это временное решение. Когда будут бабки — Named Tunnel + домен.

📋 Шпаргалка (копировать целиком)
bash
# 1. Бэкенд
cd ~/event_horizon && make deploy && curl -fsS http://localhost:8079/health

# 2. Фронтенд (терминал 1)
cd ~/event_horizon/frontend && npm run dev

# 3. Туннель (терминал 2)
pkill -f 'cloudflared tunnel --url http://localhost:5173' || true
nohup cloudflared tunnel --url http://localhost:5173 > ~/eh-frontend-tunnel.log 2>&1 &
sleep 3
grep -o 'https://[^ ]*\.trycloudflare\.com' ~/eh-frontend-tunnel.log | tail -n 1

# 4. Скопировать ссылку, Ctrl+C в терминале 1, потом:
__VITE_ADDITIONAL_SERVER_ALLOWED_HOSTS="<ССЫЛКА_ИЗ_ШАГА_3>" npm run dev

# 5. Проверить с телефона: https://<ССЫЛКА_ИЗ_ШАГА_3>

# 6. Выключить
pkill -f 'cloudflared tunnel --url http://localhost:5173'