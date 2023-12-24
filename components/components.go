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

type Complex struct {
	Target entity.Ref
}

func (p Position) Dist(p2 Position) float64 {
	return math.Sqrt(math.Pow(float64(p.X-p2.X), 2) + math.Pow(float64(p.Y-p2.Y), 2))
}
