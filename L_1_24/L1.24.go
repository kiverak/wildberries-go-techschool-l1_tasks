package main

import (
	"fmt"
	"math"
)

//Разработать программу нахождения расстояния между двумя точками на плоскости. Точки представлены в виде структуры Point
//с инкапсулированными (приватными) полями x, y (типа float64) и конструктором. Расстояние рассчитывается по формуле
//между координатами двух точек.
//
//Подсказка: используйте функцию-конструктор NewPoint(x, y), Point и метод Distance(other Point) float64.

// Point представляет точку на 2D-плоскости
type Point struct {
	x float64
	y float64
}

// NewPoint - функция-конструктор для создания нового экземпляра Point
func NewPoint(x, y float64) *Point {
	return &Point{
		x: x,
		y: y,
	}
}

// Distance вычисляет расстояние до другой точки
func (p *Point) Distance(other *Point) float64 {
	dx := p.x - other.x
	dy := p.y - other.y
	return math.Sqrt(dx*dx + dy*dy)
}

func main() {
	p1 := NewPoint(1.0, 2.0)
	p2 := NewPoint(4.0, 6.0)

	distance := p1.Distance(p2)

	fmt.Printf("Расстояние между точками: %f\n", distance)
}
