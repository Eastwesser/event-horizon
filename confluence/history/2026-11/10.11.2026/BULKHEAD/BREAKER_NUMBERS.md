authCB        := newServiceBreaker("auth", 128)
gameCB        := newServiceBreaker("game", 128)
profileCB     := newServiceBreaker("profile", 128)
leaderboardCB := newServiceBreaker("leaderboard", 128)
billingCB     := newServiceBreaker("billing", 64)
inventoryCB   := newServiceBreaker("inventory", 256)
shopCB        := newServiceBreaker("shop", 256)
paymentCB     := newServiceBreaker("payment", 64)
authorsCB     := newServiceBreaker("authors", 64)
notificationCB := newServiceBreaker("notification", 64)
historyCB     := newServiceBreaker("history", 64)
analyticsCB   := newServiceBreaker("analytics", 32)

Почему такие цифры (кратко, чтобы не забыть)
256 для shop/inventory — read path под 140 browse-VU × 4 ручки ≈ 560, но каждая ручка идёт в свой сервис: shop/items → 140 в Shop, inventory/items → 140 в Inventory. 256 — с запасом, если VU вырастет до 200.

128 для auth — пока без кэша ValidateToken (Шаг 4), Auth получает 560 rps (140 × 4). 128 — с запасом на 1 запрос = 1 ValidateToken. После Шага 4 (кэш) можно снизить до 64.

128 для leaderboard/profile — read path, но лёгкие (Redis Sorted Set, PG+Redis), 128 норм.

128 для game — submit, write path, но быстрый (outbox INSERT).

64 для billing/payment/authors — деньги и write path, не хочется 256 параллельных SpendCurrency (contention на balance, deadlock-риск на user_currencies). 64 — норм для 20 purchase-VU.

64 для notification/history — read path, но редко (не в browse-профиле).

32 для analytics — admin-only, редко.


Почему хардкод — ок
Три причины:

Это не бизнес-логика, это инфраструктурный лимит. Bulkhead — как MaxRequests: 3 и Timeout: 10s рядом. Они тоже хардкод. Если бы мы выносили все параметры в env, gateway.go стал бы помойкой из os.Getenv("SHOP_MAX_CONCURRENT") × 12. Это хуже, чем 12 чисел в одном месте.

Разные лимиты для разных сервисов — это осознанное решение, не конфиг. shop=256, billing=64 — это архитектура (read-heavy vs write-heavy), не «настройка под окружение». В prod, dev, staging — одни и те же пропорции. Если бы это был один лимит на всё — можно env. Но у нас 12 разных.

12 чисел в одном файле — легко увидеть и сравнить. grep newServiceBreaker — и видишь всю картину. Если бы это были env, пришлось бы 12 раз os.Getenv, и никто бы не увидел пропорции.

Когда стоит вынести в env: если разные окружения требуют разных лимитов (dev=32, prod=256). У нас не так — 8 GiB all-in-one и есть prod-конфиг (по README, «Prefer 8 vCPU / 32 GiB» — следующий шаг, не сейчас). Если потом понадобится — тогда вынесем. Сейчас — хардкод.

cd /home/denismatveev/event_horizon/services/gateway/internal/app
sed -i \
  -e 's/newServiceBreaker("auth")/newServiceBreaker("auth", 128)/' \
  -e 's/newServiceBreaker("game")/newServiceBreaker("game", 128)/' \
  -e 's/newServiceBreaker("profile")/newServiceBreaker("profile", 128)/' \
  -e 's/newServiceBreaker("leaderboard")/newServiceBreaker("leaderboard", 128)/' \
  -e 's/newServiceBreaker("billing")/newServiceBreaker("billing", 64)/' \
  -e 's/newServiceBreaker("inventory")/newServiceBreaker("inventory", 256)/' \
  -e 's/newServiceBreaker("shop")/newServiceBreaker("shop", 256)/' \
  -e 's/newServiceBreaker("payment")/newServiceBreaker("payment", 64)/' \
  -e 's/newServiceBreaker("authors")/newServiceBreaker("authors", 64)/' \
  -e 's/newServiceBreaker("notification")/newServiceBreaker("notification", 64)/' \
  -e 's/newServiceBreaker("history")/newServiceBreaker("history", 64)/' \
  -e 's/newServiceBreaker("analytics")/newServiceBreaker("analytics", 32)/' \
  gateway.go