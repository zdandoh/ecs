package main

import "math"

//go:generate go run github.com/zdandoh/ecs/codegen ecspkg 10000000

type Position struct {
	X int
	Y int
}

type Health struct {
	Value int
}

func (p Position) Dist(p2 Position) float64 {
	return math.Sqrt(math.Pow(float64(p.X - p2.X), 2) + math.Pow(float64(p.Y - p2.Y), 2))
}