Как искать
Открой по одному URL. Смотри — исчез ли квадрат.

Порядок поиска (от самого вероятного)
1. hide=bg-nebula — самое вероятное.

text
http://localhost:5173/?motion=force&hide=bg-nebula
Если квадрат исчез → источник — bg-void / bg-nebula на родителе. Это было моё предположение.

NO, THE SQUARE STILL EXISTS AND MOVES.

2. hide=glow-filter — второе по вероятности.

text
http://localhost:5173/?motion=force&hide=glow-filter
Если квадрат исчез → filter / backdrop-filter на glow даёт прямоугольник размытого фона.

NO, THE SQUARE STILL EXISTS AND MOVES.

3. hide=glow — сам glow.

text
http://localhost:5173/?motion=force&hide=glow

NO, THE SQUARE STILL EXISTS AND MOVES.


4. hide=hero — hero-баннер.

text
http://localhost:5173/?motion=force&hide=hero
Если квадрат исчез → hero влияет на диск (перекрытие, тень, z-index).

NO, THE SQUARE STILL EXISTS AND MOVES.

5. hide=disk — весь диск.

text
http://localhost:5173/?motion=force&hide=disk
Если квадрат исчез → источник внутри диска. Возвращаемся к отдельным слоям.

ВООБЩЕ ВСЕ ИСЧЕЗЛОТ НАХРЕН!

6. hide=pull и hide=rings и hide=core и hide=particles — по очереди.

Что делать
Открывай по одному URL. Скидывай:

URL.

Квадрат — есть / нет.

5-6 URL — 3 минуты. Определим источник.

Приоритет проверки
Сначала — hide=bg-nebula. Если квадрат исчез — это финальный ответ, и мы понимаем причину:

Родитель имеет bg-void #0b0e1a.

Page фон — чуть другой оттенок (например, более синий).

Прямоугольник родителя виден на этом фоне как квадрат.

Fix: убрать bg-void с родителя или унифицировать фон страницы до #0b0e1a.

Что я держу в карте
Закрыто (на GitHub):

v1.0.9, Wave 1, Wave 2 #1–5.

Hero, smooth scroll.

Сейчас:

🟧 Commit ebf6a92 — точка отката. ✅

🟧 Layer toggles — ?hide=... готовы.

🟧 Поиск источника квадрата — твой ход.

🟧 Rings thicker — после.

🟧 PNG crop — Photopea.

⏸ Push.

⏸ Wave 2 #6.

Отложено:

Cursor pull-in, Wave 2 #5b, Refund window.