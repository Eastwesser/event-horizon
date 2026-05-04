package hexagons

import (
    "fmt"
)

type Move struct {
    FromX int
    FromY int
    ToX   int
    ToY   int
	Timestamp int32
}

type Board struct {
    Width  int
    Height int
    Cells  [][]string  // типы плиток
}

type Validator struct{}

func NewValidator() *Validator {
    return &Validator{}
}

// ValidateMoves проверяет, что последовательность ходов валидна и ведёт к полученному счёту
func (v *Validator) ValidateMoves(seed string, moves []Move, finalScore int) (bool, int, error) {
    // TODO: эмуляция игры на сервере
    // Пока заглушка для development
    
    // Временно: пропускаем валидацию, но проверяем, что ходы не пустые
    if len(moves) == 0 && finalScore > 0 {
        return false, 0, fmt.Errorf("no moves provided but score > 0")
    }
    
    // Заглушка: считаем, что всё валидно
    return true, finalScore, nil
}
