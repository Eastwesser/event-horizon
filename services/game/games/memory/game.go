// services/game/games/memory/game.go
package memory

import (
	"math/rand"
)

type Move struct {
	CardIndex1 int `json:"card_index_1"`
	CardIndex2 int `json:"card_index_2"`
}

type MemoryGame struct {
	Board   []string `json:"board"`   // 30 карт, 15 пар
	Moves   int      `json:"moves"`
	Matched int      `json:"matched"`
	Score   int      `json:"score"`
}

const (
	MIN_SCORE = 100
	MAX_SCORE = 1000
	TOTAL_PAIRS = 15
)

// NewMemoryGame создаёт новую игру с детерминированной доской по seed
func NewMemoryGame(seed string) *MemoryGame {
	rng := rand.New(rand.NewSource(hashSeed(seed)))
	
	// 15 пар фруктов
	fruits := []string{
		"🍎", "🍒", "🍊", "🍋", "🍉",
		"🥝", "🍓", "🍑", "🥥", "🥑",
		"🍇", "🍐", "🍈", "🫐", "🍌",
	}
	
	// Создаём пары (каждый фрукт дважды)
	board := make([]string, 0, 30)
	for _, fruit := range fruits {
		board = append(board, fruit, fruit)
	}
	
	// Перемешиваем детерминированно
	rng.Shuffle(len(board), func(i, j int) {
		board[i], board[j] = board[j], board[i]
	})
	
	return &MemoryGame{
		Board:   board,
		Moves:   0,
		Matched: 0,
		Score:   0,
	}
}

// hashSeed конвертирует строковый seed в int64
func hashSeed(seed string) int64 {
	var h int64
	for _, c := range seed {
		h = 31*h + int64(c)
	}
	return h
}

// ValidateMoves проверяет валидность ходов и возвращает корректный счёт
// Для MVP: доверяем клиенту, но проверяем базовую логику
func (g *MemoryGame) ValidateMoves(seed string, moves []Move, finalScore int) (bool, int, error) {
	// Если ходы не переданы — доверяем клиенту (как в hexagon)
	if len(moves) == 0 {
		return true, finalScore, nil
	}
	
	// Восстанавливаем доску по seed
	game := NewMemoryGame(seed)
	
	// Эмулируем игру
	matched := make([]bool, 30)
	correctMoves := 0
	
	for _, move := range moves {
		// Проверяем индексы
		if move.CardIndex1 < 0 || move.CardIndex1 >= 30 ||
		   move.CardIndex2 < 0 || move.CardIndex2 >= 30 {
			return false, 0, nil
		}
		
		// Пропускаем уже найденные пары
		if matched[move.CardIndex1] || matched[move.CardIndex2] {
			continue
		}
		
		game.Moves++
		
		// Проверяем совпадение
		if game.Board[move.CardIndex1] == game.Board[move.CardIndex2] {
			matched[move.CardIndex1] = true
			matched[move.CardIndex2] = true
			game.Matched++
			correctMoves++
		}
	}
	
	// Вычисляем счёт
	calculatedScore := CalculateScore(game.Moves)
	
	// Если счёт не совпадает с клиентским — возвращаем свой
	if calculatedScore != finalScore {
		return true, calculatedScore, nil
	}
	
	return true, finalScore, nil
}

// CalculateScore вычисляет очки по формуле
func CalculateScore(moves int) int {
	idealMoves := TOTAL_PAIRS // 15 ходов = идеально
	
	if moves <= idealMoves {
		return MAX_SCORE
	}
	
	// Каждый лишний ход снижает на 20 очков
	penalty := (moves - idealMoves) * 20
	score := MAX_SCORE - penalty
	
	if score < MIN_SCORE {
		return MIN_SCORE
	}
	return score
}

// CalculateRewards вычисляет награду за игру
func (g *MemoryGame) CalculateRewards(score int) (lamps int, tickets int) {
	// База: 5 лампочек + 5 билетиков за игру
	lamps = 5
	tickets = 5
	
	// Бонус: +1 билетик за каждые 100 очков свыше 500
	if score > 500 {
		tickets += (score - 500) / 100
	}
	
	// Максимум 20 билетиков за игру
	if tickets > 20 {
		tickets = 20
	}
	
	// Бонусные лампочки за высокий результат
	if score >= 900 {
		lamps += 5
	}
	
	return lamps, tickets
}
