package main

import (
	"fmt"
	ecs "github.com/zdandoh/ecs/ecspkg"
	"testing"
)

func BenchmarkEntityCreation(b *testing.B) {
	ecs.Reset()

	for n := 0; n < b.N; n++ {
		ecs.NewEntity()
	}
}

func BenchmarkEntityCreateRemove(b *testing.B) {
	ecs.Reset()

	for n := 0; n < b.N; n++ {
		e := ecs.NewEntity()
		e.Kill()
	}
}

func BenchmarkComponentCreation(b *testing.B) {
	ecs.Reset()

	for n := 0; n < b.N; n++ {
		e := ecs.NewEntity()
		e.AddPosition(ecs.Position{3, 4})
	}
}

func BenchmarkComponentSelect(b *testing.B) {
	ecs.Reset()

	for i := 0; i < 10_000; i++ {
		e := ecs.NewEntity()
		e.AddHealth(ecs.Health{54})
	}
	for i := 0; i < 100; i++ {
		e := ecs.NewEntity()
		e.AddPosition(ecs.Position{32, 45})
	}

	found := 0

	for n := 0; n < b.N; n++ {
		found = 0
		ecs.SelectWithComponent(func(e ecs.Entity) {
			found += 1
		}, ecs.ComponentHealth)
	}

	fmt.Println(found)
}

func BenchmarkComponentMultiSelect(b *testing.B) {
	ecs.Reset()

	for i := 0; i < 100_000; i++ {
		e := ecs.NewEntity()
		e.AddHealth(ecs.Health{54})
	}
	for i := 0; i < 100; i++ {
		e := ecs.NewEntity()
		e.AddPosition(ecs.Position{32, 45})
		e.AddHealth(ecs.Health{21})
	}

	found := 0
	for n := 0; n < b.N; n++ {
		ecs.SelectWithComponents(func(e ecs.Entity) {
			found += 1
		}, ecs.ComponentHealth, ecs.ComponentPosition)
	}
}

func TestECS(t *testing.T) {
	ecs.Reset()

	for i := 0; i < 100; i++ {
		e := ecs.NewEntity()
		e.AddHealth(ecs.Health{45})
	}
	for i := 0; i < 100; i++ {
		e := ecs.NewEntity()
		e.AddPosition(ecs.Position{23, 45})
	}

	ecs.SelectWithComponent(func(e ecs.Entity) {
		e.Health().Value += 1
	}, ecs.ComponentHealth)

	ecs.SelectWithComponent(func(e ecs.Entity) {

	}, ecs.ComponentPosition)

	ecs.SelectWithComponent(func(e ecs.Entity) {
		if e.Health().Value != 46 {
			t.FailNow()
		}
	}, ecs.ComponentHealth)
}
