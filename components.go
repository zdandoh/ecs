package main

//go:generate go run ecs/codegen ecs 10000000

type Health struct {
	Value int
}

type Position struct {
	X int
	Y int
}
