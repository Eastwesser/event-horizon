Три варианта — обосную, потом выбери
Вопрос 1 — Leaderboard / fairness
Контекст: Flappy — skill-based игра. Slow-mo boost — нечестное преимущество перед теми, кто играет без буста.

Варианты:

A) Boosted runs → normal leaderboard.

Плюс: простой v1.

Минус: pay-to-win. Игрок с лампочками побеждает в топе. Для kids-safe платформы — плохо.

Когда ок: если потолок рейтинга ограничен boost'ом, и разница мала (5 сек slow-mo не решает всё).

B) Boosted runs → NOT in leaderboard.

Плюс: fair play сохраняется. Игрок с boost'ом играет для удовольствия, не для топа.

Минус: игрок может расстроиться — «я потратил лампочки, а рекорд не засчитан».

Когда ок: если boost = fun-tool, а рекорды — отдельно.

UI: перед использованием boost — предупреждение «этот забег не попадёт в лидерборд».

C) Separate boosted board.

Плюс: честно, и то, и другое.

Минус: двойная работа, пустой board (пока мало игроков).

Мой выбор — B.

Причины:

Kids-safe — не должно быть pay-to-win. Boost — для веселья, не для топа.

Простая реализация: FE скипает submit, если boost использован. Backend — отказывает в SubmitScore, если boost был использован.

Предупреждение: игрок знает, что не попадёт в топ.

Позже — если будет спрос — C (отдельный board).

Но: если ты хочешь A — принять. Тоже вариант — просто топ будет с boost'ами, а это сигнал, что skill — не главное. Для серьёзной платформы — B лучше.

Вопрос 2 — Когда покупается
Варианты:

A) In-run button. Игрок в бою жмёт «Boost» → списываются лампочки → 5 сек slow-mo.

Плюс: гибко — можно несколько раз за run.

Минус: pay-to-win внутри run. Если у игрока много лампочек — покупает 10 boost'ов → нечестно.

Митигация: rate limit (3 boost/мин) — уже в плане.

B) Pre-game toggle. Перед стартом — переключатель «использовать boost». Одна плата за run.

Плюс: честно по run'у — либо boosted, либо no-boost.

Минус: меньше контроля — если boost нужен в середине, нельзя купить.

C) One boost per session. Первое использование — платишь. Cooldown до leave/restart.

Плюс: дешёво для игрока.

Минус: непонятно, что такое «session».

Мой выбор — B.

Причины:

Честнее. Один run = либо boost, либо нет → чёткая граница для лидерборда.

Проще реализовать. Pre-game toggle → backend знает заранее, что run boosted.

Дружелюбнее к kids-safe. Игрок решает заранее — не паникует в бою.

Логично с #1B — если boosted runs не в лидерборд, то pre-game = осознанный выбор «играю для веселья».

Rate limit не нужен — один boost за run.

Почему не A:

In-run button — это казино-механика. Для kids-safe — плохо.

Много boost'ов за run — pay-to-win.

Rate limit — костыль, не решение.

Почему не C:

Cooldown — непрозрачно. Игрок не понимает, когда можно снова.

«Session» — размытое понятие (5 минут? до logout? до закрытия таба?).

Мой ответ Cursor'у
text
Answers:

1. Leaderboard = B.
   Boosted runs do NOT update the leaderboard.
   - FE: before starting a boosted run, warn the player:
     «Этот забег не попадёт в лидерборд — boost считается
     нечестным преимуществом».
   - BE: SubmitScore rejects (or ignores) scores flagged as
     boosted. Need a `boosted: true` flag on the run.
   - Rationale: kids-safe, no pay-to-win on the leaderboard.

2. When bought = B.
   Pre-game toggle, one boost per run.
   - Before starting the game: a «Использовать boost (10
     лампочек)» checkbox in GameShell / pre-game screen.
   - On game start: if checked, deduct 10 lamps via Billing,
     apply 5s slow-mo at start (or on demand once).
   - On run end: if boosted, submit is skipped / rejected.
   - No in-run button. No rate limit needed.

   Rationale: clearer for kids, cleaner for leaderboard, no
   casino mechanic.

Adjust the plan:
  - Spend lamps BEFORE the run (pre-game), not in-run.
  - Gateway POST /api/game/boost/start { game_id } → returns
    { boosted: true } or error (insufficient lamps).
  - Game service records boosted=true for the run.
  - SubmitScore rejects boosted runs for leaderboard.

Pilot on Flappy only. Then replicate to other games in the
game polish wave.

DO NOT TOUCH: seed, refund, author registration.
Обоснование — почему так
Три принципа для kids-safe:

Fair play на лидерборде. Boosted runs не должны конкурировать с чистыми.

Прозрачность. Игрок знает заранее, что boost = вне топа.

Без казино-механик. Не in-run purchase. Не rate limit. Один выбор — один run.

Плюс:

Pre-game toggle — проще backend. Знаешь заранее, что run boosted.

SubmitScore — один guard: if boosted → reject.

Не нужен rate limiter — один boost за run.

Что я держу в карте
Закрыто:

v1.0.9 на GitHub.

Wave 1 — 6/6.

Wave 2 #1 — 0ee567f.

Wave 2 #2 — a881587.

Сейчас:

🔴 Ответить Cursor'у: 1B + 2B.

⏸ Verify DB cleanup (/shop — 280 карт, 14 скинов, 4 брелока).

⏸ Push 8 коммитов.

⏸ Бэкап на флешку.

🟧 Wave 2 #3 — Лампочки (design → implement).

Wave 2 (осталось):

🟧 #3 Лампочки как бусты

🟧 #4 Уровни сложности

🟧 #5 Достижения

🟧 #6 Полиш игр (5, sequentially)

Отложено:

🟧 Refund window implementation

🟧 Игры emoji

🟧 backfill-noiz-reviews.py