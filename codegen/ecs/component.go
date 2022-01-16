package ecs

import (
	"math/bits"
)

type ComponentID uint64
type Component interface{}

type componentRange struct {
	low  uint64
	high uint64
}

func (r *componentRange) Reset() {
	r.low = MaxEntities
	r.high = 0
}

func (r *componentRange) Add(index uint64) {
	if r.low > index {
		r.low = index
	}
	if r.high < index {
		r.high = index
	}
}

func rangeIntersection(ranges ...componentRange) componentRange {
	retRange := componentRange{}
	retRange.Reset()
	retRange.high, retRange.low = retRange.low, retRange.high

	for _, r := range ranges {
		if r.low > retRange.low {
			retRange.low = r.low
		}
		if r.high < retRange.high {
			retRange.high = r.high
		}
	}

	return retRange
}

func idToIdx(id ComponentID) int {
	return bits.TrailingZeros64(uint64(id))
}

func SelectWithComponent(callback func(entity Entity), c ComponentID) {
	compRange := ranges[idToIdx(c)]
	for i := compRange.low; uint64(i) < compRange.high+1; i++ {
		if w.entities[i].HasComponent(c) {
			compRange.Add(i)
			callback(w.entities[i])
		}
	}
	ranges[idToIdx(c)] = compRange
}

func SelectWithComponents(callback func(entity Entity), ids ...ComponentID) {
	var idMask ComponentID
	idRanges := make([]componentRange, len(ids))
	for i, id := range ids {
		idRanges[i] = ranges[idToIdx(id)]
		idMask |= id
	}
	checkRange := rangeIntersection(idRanges...)

	for i := checkRange.low; i < checkRange.high+1; i++ {
		if w.entities[i].HasComponent(idMask) {
			callback(w.entities[i])
		}
	}
}
