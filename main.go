package main

import (
	"ecs/ecs"
	"fmt"
)

func main() {
	e := ecs.NewEntity()
	e.AddPosition(ecs.Position{3, 4})
	e.AddHealth(ecs.Health{30})

	fmt.Println(e.Position())
}
