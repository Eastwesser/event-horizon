# MEMONIA (rulebook)

# 📖 Как считаются очки:
- Идеально: 15 ходов → 1000 очков
- Каждый лишний ход → -20 очков
- Минимум: 100 очков

# 🎯 Совет: запоминай, где лежат парные карты!


# Множители (баффы)

Combo	Множитель	Эффект
0-1	    1x	        обычные очки
2-3	    2x	        удвоение
4+	    3x	        утроение (ура!)

# Анимация переворота

```css
.memory-card {
  transition: transform 0.5s;
  transform-style: preserve-3d;
}

.memory-card.flipped {
  transform: rotateY(180deg);
}

.memory-card.matched {
  opacity: 0;
  pointer-events: none;
  transition: opacity 0.3s ease;
}
```

# Счётчик ходов → очки в лидерборд

```ts
interface MemoryState {
  cards: Card[];           // 30 карт, 15 пар
  flippedIndices: number[]; // максимум 2
  matchedPairs: number;     // счёт найденных
  moves: number;            // количество ходов
  combo: number;            // подряд правильные пары
  multiplier: number;       // 1x, 2x, 3x
  gameOver: boolean;
  
  flipCard: (index: number) => void;
  checkMatch: () => void;
  resetGame: () => void;
}
```

# Формула очков (прозрачная и честная):

```typescript
// Максимум очков за идеальную игру: 1000
// Минимум: 100 (если очень долго искал)

function calculateScore(moves: number, totalPairs: number): number {
  const MIN_SCORE = 100;
  const MAX_SCORE = 1000;
  const IDEAL_MOVES = totalPairs; // 15 ходов = идеально (каждая пара с первого раза)
  
  let score = MAX_SCORE;
  if (moves > IDEAL_MOVES) {
    // Каждый лишний ход снижает очки
    const penalty = Math.min(MAX_SCORE - MIN_SCORE, (moves - IDEAL_MOVES) * 20);
    score = MAX_SCORE - penalty;
  }
  
  return Math.max(MIN_SCORE, score);
}
```

# Фрукты для 15 пар:

```typescript
const FRUIT_EMOJIS = [
  '🍎', '🍒', '🍊', '🍋', '🍉',
  '🥝', '🍓', '🍑', '🥥', '🥑',
  '🍇', '🍐', '🍈', '🫐', '🍌'
];
```

# Сетка: 3 ряда × 10 карт (30 карт)

- На десктопе: flex-wrap с фиксированной шириной карты
- На мобилке: скролл или уменьшенный размер

```css
.memory-board {
  display: grid;
  grid-template-columns: repeat(10, 1fr);
  gap: 10px;
  max-width: 1000px;
  margin: 0 auto;
}

.memory-card {
  aspect-ratio: 1 / 1;
  max-width: 80px;
  margin: 0 auto;
}
```

Про 200 очков 🎯

Логи:
score=200, lamps=5, tickets=5

Ты получил 200 очков за игру, где нашёл все пары. Формула работает так:
Идеально: 15 ходов → 1000 очков
У тебя было больше ходов → очки снизились до 200

Комбо (x2, x3) влияет на впечатление, но не на финальные очки. 
Очки считаются только по формуле 1000 - (лишние ходы × 20). 

Комбо — это просто визуальный бафф, чтобы игрок чувствовал прогресс. 
Никакого скама! Всё честно и прозрачно.
