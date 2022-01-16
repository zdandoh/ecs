package ecs

type Entity struct {
	id         uint64
	generation uint64
	alive      bool
	components ComponentID
}

func (e *Entity) Kill() {
	w.freeList = append(w.freeList, e.id)
	e.alive = false
}

func (e Entity) HasComponent(comp ComponentID) bool {
	return e.components&comp == comp
}
