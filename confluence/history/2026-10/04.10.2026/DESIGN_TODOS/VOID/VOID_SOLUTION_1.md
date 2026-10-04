Нашли. reducedMotion: true
Что это значит
text
[VOID] mount {
  dev: true
  href: "http://localhost:5173/"
  mode: "development"
  reducedMotion: true    ← ВОТ ОНО
}
prefers-reduced-motion: reduce активен. Наша фича намеренно выключает все анимации — правильно, для accessibility.

Ты не видишь анимации, потому что система считает, что ты попросил уменьшить движение.

Это не баг. Это фича для людей с вестибулярными расстройствами, эпилепсией, чувствительностью к движению.

Почему true у тебя
Варианты:

Windows → Settings → Accessibility → Visual effects → Animation effects — выключено.

Chrome → флаг / расширение — эмулирует reduce.

Виртуалка — virtual GPU сообщает reduce.

DevTools → Rendering → Emulate prefers-reduced-motion — редуцировано (по умолчанию No emulation — но проверить).

Как выключить — 4 места
1. DevTools → Rendering
F12 → вкладка Rendering (может быть скрыта — в трёх точках »).

Найди Emulate CSS media feature prefers-reduced-motion:

Должно быть — No emulation.

Если стоит reduce — поменяй на No emulation.

Refresh. Hover.

2. Windows Settings
Settings → Accessibility → Visual effects → Animation effects — On.

Или старый путь — Settings → Ease of Access → Display → Show animations in Windows — On.

3. Chrome флаги
chrome://flags → поиск reduced-motion → не должно быть force.

4. Система — Win + R → sysdm.cpl
Advanced → Performance → Settings → Visual Effects — Let Windows choose или Custom с галочкой Animate controls and elements inside windows.

Что проверить
После изменения одного из этих — refresh, hover.

Смотри в Console:

text
[VOID] mount { ..., reducedMotion: false }
Если false — анимация заработает.

Если всё ещё true — что-то ещё эмулирует. Проверь:

js
window.matchMedia('(prefers-reduced-motion: reduce)').matches
В Console. Если true — система редуцирует.

Быстрый тест — force disable
В DevTools → Console вставь:

js
// override — на время сессии
window.matchMedia('(prefers-reduced-motion: reduce)').matches
Если true — проблема на уровне системы / браузера.

Или — открой Home с параметром (если Cursor добавит):

js
// например ?nomotion=1 → forces animations on
Но это — Cursor'у отдельно.

Промпт Cursor'у
text
Found it: reducedMotion: true in the mount payload.

The system / browser reports prefers-reduced-motion: reduce,
so the code intentionally kills all disk animations. Not a
bug — the feature is working as designed.

Two things:

1. Add a dev override for testing:
   - If `?motion=force` is in URL → ignore prefers-reduced-motion
     and always animate.
   - This lets me test the visuals without changing OS settings.
   - Not a production feature; dev-only.
   - Gate it behind import.meta.env.DEV.

2. Keep prefers-reduced-motion behavior for real users —
   it's correct for accessibility.

DO NOT TOUCH: hero, nav, games, seed, refund, authors.
No push.
Что делать тебе
Быстрое решение (30 сек)
DevTools → Rendering → Emulate prefers-reduced-motion → No emulation.

Refresh. Hover. Работает?

Или (если DevTools не помог)
Windows Settings → Accessibility → Animation effects → On.

Refresh. Hover. Работает?

Или — попроси Cursor'а сделать override
Промпт выше — ?motion=force для dev.

Что я держу в карте
Закрыто (на GitHub):

v1.0.9, Wave 1, Wave 2 #1–5.

Hero, smooth scroll.

Сейчас:

🟧 VOID v3 — 6722de1 (local).

✅ Диагноз: reducedMotion: true — наша фича выключила анимации.

🟧 Выключить prefers-reduced-motion на твоей машине → или ?motion=force.

🟧 Wordmark flush — Photopea crop.

⏸ Push.

⏸ Wave 2 #6.

Отложено:

Cursor pull-in, Wave 2 #5b, Refund window.

