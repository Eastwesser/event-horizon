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

// type Board struct {
//     Width  int
//     Height int
//     Cells  [][]string  // типы плиток
// }

type Validator struct{}

func NewValidator() *Validator {
    return &Validator{}
}

// // ValidateMoves — полная валидация игры
// func (v *Validator) ValidateMoves(seed string, moves []Move, finalScore int) (bool, int, error) {
//     // Восстанавливаем доску
//     fmt.Printf("🔍 ValidateMoves: seed=%s, moves=%d\n", seed, len(moves))
//     board, err := GenerateInitialBoard(seed, 3)
//     if err != nil {
//         return false, 0, err
//     }
//     fmt.Printf("🔍 Initial board generated\n")
    
//     // Применяем ходы
//     for i, move := range moves {
//         from := Coord{Q: move.FromX, R: move.FromY}
//         to := Coord{Q: move.ToX, R: move.ToY}
//         fmt.Printf("🔍 Applying move %d: (%d,%d) -> (%d,%d)\n", i, from.Q, from.R, to.Q, to.R)

//         if err := board.MoveTile(from, to); err != nil {
//             return false, 0, fmt.Errorf("move %d invalid: %w", i, err)
//         }
//     }

//     // Вычисляем счёт
//     calculatedScore := board.CalculateScore()
//     fmt.Printf("🔍 Calculated score: %d\n", calculatedScore)

//     // finalScore игнорируем — сервер сам считает
//     return true, calculatedScore, nil
// }

func (v *Validator) ValidateMoves(seed string, moves []Move, finalScore int) (bool, int, error) {
    board, err := GenerateInitialBoard(seed, 3)
    if err != nil {
        return false, 0, err
    }
    
    for i, move := range moves {
        from := Coord{Q: move.FromX, R: move.FromY}
        to := Coord{Q: move.ToX, R: move.ToY}
        if err := board.MoveTile(from, to); err != nil {
            return false, 0, fmt.Errorf("move %d invalid: %w", i, err)
        }
    }
    
    calculatedScore := board.CalculateScore()
    fmt.Printf("🔍 Final score: %d\n", calculatedScore)
    
    return true, calculatedScore, nil
}