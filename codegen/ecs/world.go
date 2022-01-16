package PACKAGE_NAME

type worldState struct {
	compCount uint64
	entCount  uint64
	freeList  []uint64
	entities  []Entity
	systems   []System
}

const allowOverflow = true

var w worldState

func init() {
	w.entities = make([]Entity, MaxEntities)
}

func Reset() {
	w = worldState{}
	w.entities = make([]Entity, MaxEntities)
	clearComponents()
}

func NewEntity() *Entity {
	if allowOverflow && w.entCount >= MaxEntities {
		w.entCount = 0
	}

	var currIndex uint64
	if len(w.freeList) == 0 {
		currIndex = w.entCount
		w.entCount++
	} else {
		currIndex = w.freeList[len(w.freeList)-1]
		w.freeList = w.freeList[:len(w.freeList)-1]
	}
	w.entities[currIndex] = Entity{currIndex, w.entities[currIndex].generation + 1, true, 0}

	return &w.entities[currIndex]
}

func AddSystem(s System) {
	w.systems = append(w.systems, s)
}

func Update() {
	for _, s := range w.systems {
		s.Update()
	}
}
