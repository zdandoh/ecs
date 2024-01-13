package components

import (
	"math"

	"github.com/zdandoh/ecs/ecspkg/entity"
)

type Velocity struct {
	X int
	Y int
}

type Health int

type Position struct {
	X int
	Y int
}

type Pos struct {
	X float64
	Y float64
}

type Vel struct {
	X float64
	Y float64
}

type Complex struct {
	Target entity.Ref
}

func (p Position) Dist(p2 Position) float64 {
	return math.Sqrt(math.Pow(float64(p.X-p2.X), 2) + math.Pow(float64(p.Y-p2.Y), 2))
}

type Has struct {
	Relationship struct{}
	Count        int
}

type Likes struct {
	Relationship struct{}
}
