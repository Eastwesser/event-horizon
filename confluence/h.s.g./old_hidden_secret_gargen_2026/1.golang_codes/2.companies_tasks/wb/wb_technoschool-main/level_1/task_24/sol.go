package main

import (
	"fmt"
	"math"
)

type Point struct{ x, y float64 }

func NewPoint(x, y float64) Point { return Point{x, y} }

func (p Point) Distance(q Point) float64 {
	dx := p.x - q.x
	dy := p.y - q.y
	return math.Hypot(dx, dy)
}

func main() {
	a := NewPoint(0, 0)
	b := NewPoint(3, 4)
	fmt.Println(a.Distance(b)) // 5
}
