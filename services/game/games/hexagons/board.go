package hexagons

import (
    "fmt"
    "math/rand"
)

// Coord — кубические координаты гексагона
type Coord struct {
    Q, R int
}

// String для отладки
func (c Coord) String() string {
    return fmt.Sprintf("(%d,%d)", c.Q, c.R)
}

// Neighbors возвращает все 6 соседних координат
func (c Coord) Neighbors() []Coord {
    return []Coord{
        {c.Q - 1, c.R - 1}, // верхний левый
        {c.Q, c.R - 1},     // верхний правый
        {c.Q - 1, c.R},     // левый
        {c.Q + 1, c.R},     // правый
        {c.Q, c.R + 1},     // нижний левый
        {c.Q + 1, c.R + 1}, // нижний правый
    }
}

// Tile — плитка на доске
type Tile struct {
    Type  string  // "twix", "oreo", "snickers", "empty"
    Coord Coord
    Count int     // 👈 количество блинов в стопке
}

// Board — гексагональная доска
type Board struct {
    Width      int
    Height     int
    Tiles      map[Coord]*Tile
    TotalScore int   // 👈 общий счёт за игру
}

// NewBoard создаёт доску заданного размера
func NewBoard(width, height int) *Board {
    board := &Board{
        Width:  width,
        Height: height,
        Tiles:  make(map[Coord]*Tile),
        TotalScore: 0,
    }
    return board
}

// SetTile устанавливает плитку на координату
func (b *Board) SetTile(q, r int, tileType string) {
    coord := Coord{Q: q, R: r}
    
    // Если плитка уже существует, сохраняем её Count
    existing := b.Tiles[coord]
    count := 1 // новая стопка начинается с 1 блина
    if existing != nil {
        count = existing.Count
    }
    
    b.Tiles[coord] = &Tile{
        Type:  tileType,
        Coord: coord,
        Count: count,
    }
}

// GetTile возвращает плитку по координате
func (b *Board) GetTile(q, r int) *Tile {
    coord := Coord{Q: q, R: r}
    return b.Tiles[coord]
}

// IsValidMove проверяет, можно ли переместить плитку с from на to
func (b *Board) IsValidMove(from, to Coord) bool {
    // 1. Проверяем, что from и to в пределах доски
    if from.Q < 0 || from.Q >= b.Width || from.R < 0 || from.R >= b.Height {
        return false
    }
    if to.Q < 0 || to.Q >= b.Width || to.R < 0 || to.R >= b.Height {
        return false
    }

    // 2. Проверяем, что from не пустая
    fromTile := b.GetTile(from.Q, from.R)
    if fromTile == nil || fromTile.Type == "empty" {
        return false
    }

    // 3. Проверяем, что to — пустая или в пределах
    toTile := b.GetTile(to.Q, to.R)
    if toTile != nil && toTile.Type != "empty" {
        return false // цель занята
    }

    // 4. Проверяем, что to — сосед from
    isNeighbor := false
    for _, neighbor := range from.Neighbors() {
        if neighbor == to {
            isNeighbor = true
            break
        }
    }
    if !isNeighbor {
        return false
    }

    return true
}

// MoveTile перемещает плитку и проверяет, не набралось ли 10
func (b *Board) MoveTile(from, to Coord) error {
    fmt.Printf("🔷 MoveTile: from (%d,%d) to (%d,%d)\n", from.Q, from.R, to.Q, to.R)

    if !b.IsValidMove(from, to) {
        return fmt.Errorf("invalid move: %v -> %v", from, to)
    }

    fromTile := b.GetTile(from.Q, from.R)
    if fromTile == nil {
        return fmt.Errorf("no tile at source")
    }
    
    // Перемещаем стопку
    movingCount := fromTile.Count
    b.SetTile(to.Q, to.R, fromTile.Type)
    
    // Увеличиваем счёт на целевой клетке
    toTile := b.GetTile(to.Q, to.R)
    if toTile != nil {
        toTile.Count = movingCount
    }
    
    b.SetTile(from.Q, from.R, "empty")

    // Проверяем стопку на целевой клетке
    toTile = b.GetTile(to.Q, to.R)
    if toTile != nil && toTile.Type != "empty" {
        totalCount := toTile.Count
        neighbors := to.Neighbors()
        for _, neighbor := range neighbors {
            neighborTile := b.GetTile(neighbor.Q, neighbor.R)
            if neighborTile != nil && neighborTile.Type != "empty" && neighborTile.Type == toTile.Type {
                totalCount += neighborTile.Count
                b.SetTile(neighbor.Q, neighbor.R, "empty")
            }
        }
        
        // Обновляем стопку с новым суммарным Count
        toTile.Count = totalCount
        
        // Если стопка достигла 10 или больше
        if totalCount >= 10 {
            b.TotalScore += totalCount
            fmt.Printf("🎉 Cleared stack of %d! Total score: %d\n", totalCount, b.TotalScore)
            b.SetTile(to.Q, to.R, "empty")
        } else {
            // Обновляем Count в tile
            b.SetTile(to.Q, to.R, toTile.Type)
            toTile.Count = totalCount
        }
    }
    
    fmt.Printf("🔷 Current TotalScore: %d\n", b.TotalScore)
    return nil
}

// GenerateInitialBoard генерирует начальное состояние (детерминированно по seed)
func GenerateInitialBoard(seed string, level int) (*Board, error) {
    // Конвертируем seed в int64
    var seedInt int64
    for _, c := range seed {
        seedInt += int64(c)
    }
    
    maxAttempts := 5
    for attempt := 0; attempt < maxAttempts; attempt++ {
        // Используем seed + attempt для разных генераций
        rng := rand.New(rand.NewSource(seedInt + int64(attempt)))
        
        // Размер доски зависит от уровня
        width := 5 + level/2
        height := 5 + level/2
        
        board := NewBoard(width, height)
        
        // 👇 1. Сначала заполняем ВСЕ клетки как "empty"
        for q := 0; q < width; q++ {
            for r := 0; r < height; r++ {
                board.SetTile(q, r, "empty")
            }
        }

        // Типы плиток (блинчики!)
        tileTypes := []string{
            "nutella", "strawberry", "fish", "sausage",
            "chicken", "caesar", "cranberry", "pancake",
        }
        
        // Количество плиток зависит от уровня (сложнее = больше плиток)
        // Уровень 1: 30% клеток, Уровень 5: 70% клеток
        fillRatio := 0.3 + float64(level-1)*0.1
        numTiles := int(float64(width*height) * fillRatio)
        
        // 2. Заполняем случайные клетки реальными плитками
        for i := 0; i < numTiles; i++ {
            q := rng.Intn(width)
            r := rng.Intn(height)
            tileType := tileTypes[rng.Intn(len(tileTypes))]
            board.SetTile(q, r, tileType)
        }
        
        // 3. Проверяем solvability
        if board.IsSolvable() {
            fmt.Printf("✅ Solvable board generated on attempt %d\n", attempt)
            return board, nil
        }
        
        fmt.Printf("Attempt %d: NOT solvable, retrying...\n", attempt)
    }
    
    return nil, fmt.Errorf("failed to generate solvable board after %d attempts", maxAttempts)
}

// Clone создаёт копию доски
func (b *Board) Clone() *Board {
    newBoard := NewBoard(b.Width, b.Height)
    for coord, tile := range b.Tiles {
        newBoard.Tiles[coord] = &Tile{
            Type:  tile.Type,
            Coord: coord,
        }
    }
    return newBoard
}

// ApplyMoves применяет последовательность ходов к доске
func (b *Board) ApplyMoves(moves []Move) error {
    board := b.Clone()
    for _, move := range moves {
        from := Coord{Q: move.FromX, R: move.FromY}
        to := Coord{Q: move.ToX, R: move.ToY}
        if err := board.MoveTile(from, to); err != nil {
            return err
        }
    }
    // Обновляем оригинальную доску
    b.Tiles = board.Tiles
    return nil
}

// IsSolvable проверяет, есть ли хотя бы один возможный ход
func (b *Board) IsSolvable() bool {
    for coord, tile := range b.Tiles {
        if tile.Type == "empty" {
            continue
        }
        for _, neighbor := range coord.Neighbors() {
            // Проверяем, что сосед существует в пределах доски
            if neighbor.Q < 0 || neighbor.Q >= b.Width || neighbor.R < 0 || neighbor.R >= b.Height {
                continue
            }
            neighborTile := b.GetTile(neighbor.Q, neighbor.R)
            // Если сосед пустой (тип "empty") — можно двигать
            if neighborTile != nil && neighborTile.Type == "empty" {
                return true
            }
        }
    }
    return false
}

// // CalculateScore вычисляет очки (за совпадающие группы)
// func (b *Board) CalculateScore() int {
//     visited := make(map[Coord]bool)
//     totalScore := 0

//     for coord, tile := range b.Tiles {
//         if tile.Type == "empty" || visited[coord] {
//             continue
//         }

//         // Находим группу одинаковых плиток
//         group := b.findGroup(coord, tile.Type)
//         for _, c := range group {
//             visited[c] = true
//         }

//         // Очки за группу: размер группы в квадрате
//         if len(group) >= 2 {
//             totalScore += len(group) * len(group)
//         }
//     }
//     fmt.Printf("🔢 CalculateScore: totalScore = %d\n", totalScore) // 👈 отладка
//     return totalScore
// }

// CalculateScore возвращает общий счёт за игру
func (b *Board) CalculateScore() int {
    fmt.Printf("🔍 CalculateScore: TotalScore = %d\n", b.TotalScore)
    return b.TotalScore
}

// findGroup находит все связанные плитки одного типа (BFS)
func (b *Board) findGroup(start Coord, tileType string) []Coord {
    group := []Coord{start}
    queue := []Coord{start}
    visited := map[Coord]bool{start: true}

    for len(queue) > 0 {
        current := queue[0]
        queue = queue[1:]

        for _, neighbor := range current.Neighbors() {
            neighborTile := b.GetTile(neighbor.Q, neighbor.R)
            if neighborTile != nil && neighborTile.Type == tileType && !visited[neighbor] {
                visited[neighbor] = true
                group = append(group, neighbor)
                queue = append(queue, neighbor)
            }
        }
    }

    return group
}
