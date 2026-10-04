Что говорит console
text
[VOID] mount { ... }
[VOID] mount { ... }       ← дважды (StrictMode)
...
[VOID] enter eh-disk eh-disk--near eh-disk--pull    ← HOVER работает!
[VOID] leave eh-disk eh-disk--near                  ← LEAVE работает!
[VOID] pull transition run none                     ← !!!
Ключевая строка:

text
[VOID] pull transition run none
Что это значит: transitionrun fired, но transition-property = none. То есть у элемента нет transition для этого свойства. Анимация не запускается — браузеру нечего анимировать.

Плюс:

enter → класс добавляется (eh-disk--near eh-disk--pull). ✅

leave → класс снимается (eh-disk). ✅

Класс приходит, но transition не срабатывает.

Дополнительный симптом
text
[VOID] pull transition run none
[VOID] pull transition run none
[VOID] pull transition run none
...
Много раз за секунду. Значит transitionrun fired, но none — браузер говорит: «нет transition для этого свойства».

Это или:

CSS-класс eh-disk--pull не определяет transition.

Или — определяет, но только для других свойств (не для того, что меняется).

Или — stale CSS — HMR не применил финальные правки.

Что делать
1. Проверь CSS — .eh-disk--pull
DevTools → Elements → наведи на .eh-disk (когда класс активен).

Computed:

transition-property — что?

animation-name — что?

Expect (что должно быть):

css
.eh-disk--pull .eh-disk-ring {
  transition: transform 1.5s ease-out;
  transform: rotateX(52deg);
}
Если transition-property: none — CSS не долетел.

2. Проверь theme.css в проекте
Cursor пишет про update theme.css — открой:

bash
grep -A5 "eh-disk--pull" frontend/src/styles/theme.css
Что там для .eh-disk--pull?

Есть transition?

Есть transform?

Скинь вывод — разберём.

3. HMR — долетел ли CSS
В консоли много [VOID] pull transition run none — и это повторяется.

Возможно — HMR подтянул JS, но CSS — stale. Или — theme.css не обновился в браузере.

Fix:

Hard refresh (Ctrl+Shift+R) ещё раз.

Или — закрой вкладку, открой заново.

Или — npm run dev → рестарт.

Cursor написал:

Your Cursor terminal still has npm run dev on 5173.

Проверь, что npm run dev видит изменения:

DevTools → Sources → найди theme.css → есть секция .eh-disk--pull?

Если нет — CSS stale. Рестарт npm run dev.

Промпт Cursor'у
text
Found it. The console shows:

  [VOID] enter eh-disk eh-disk--near eh-disk--pull
  [VOID] leave eh-disk eh-disk--near
  [VOID] pull transition run none    ← repeated many times

Classes ARE applied on enter/leave. But the transition is
running with `transition-property: none` — no actual
transition to animate.

So the CSS that should define the transition on
.eh-disk--pull is either not present, not applied, or
overridden.

Please check and report:

1. In theme.css, what exactly does .eh-disk--pull define?
   Show the rule.
2. In DevTools → Elements → .eh-disk-ring (while class is
   active), what is the computed transition-property?
3. Is there another rule overriding it (transform: none,
   transition: none, or animation: none)?
4. Is theme.css actually updated in my dev session — i.e.
   does the source show eh-disk--pull rules, or is it stale?

Report before changing anything. No push.
Что проверить тебе прямо сейчас
DevTools (F12):

Elements → наведи на диск → выбери .eh-disk-ring.

Computed → filter на transition:

transition-property: ?

transition-duration: ?

Computed → filter на transform:

transform: ?

Скинь — что видишь.

Если transition-property: none — CSS не тот.
Если transition-property: transform — CSS есть, но transform не меняется.

Что я держу в карте
Закрыто (на GitHub):

v1.0.9, Wave 1, Wave 2 #1–5.

Hero, smooth scroll.

Сейчас:

🟧 VOID v3 — 6722de1 (local).

🔴 CSS transition не применяется — transition-property: none.

🟧 Wordmark flush — Photopea crop.

⏸ Push.

⏸ Wave 2 #6.

Отложено:

Cursor pull-in, Wave 2 #5b, Refund window.