package ecs

type System interface {
	Update()
	Priority() int
}
