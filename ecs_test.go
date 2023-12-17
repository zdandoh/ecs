package main

import (
	"github.com/zdandoh/ecs/components"
	ecs "github.com/zdandoh/ecs/ecspkg"
	"testing"
)

func BenchmarkEntityCreation(b *testing.B) {
	ecs.Reset()

	b.ResetTimer()
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
		e.AddPosition(components.Position{3, 4})
	}
}

func BenchmarkSelectMatch(b *testing.B) {
	ecs.Reset()

	for i := 0; i < 10000; i++ {
		dog := ecs.NewEntity()
		dog.AddPosition(components.Position{100, 100})
		dog.AddHealth(456)
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		c := 0
		ecs.Select(func(entity ecs.Entity, health *components.Health, pos *components.Position) {
			c++
		})
		if c < 10000 {
			b.Fatal("not enough entities found")
		}
	}
}

func BenchmarkSelectUnmatched(b *testing.B) {
	ecs.Reset()

	for i := 0; i < 10000; i++ {
		dog := ecs.NewEntity()
		dog.AddPosition(components.Position{100, 100})
		dog.AddHealth(456)
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		c := 0
		ecs.Select(func(entity ecs.Entity, v *components.Velocity) {
			c++
		})
		if c > 0 {
			b.Fatal("too many entities found")
		}
	}
}

func TestHasComponent(t *testing.T) {
	ecs.Reset()

	dog := ecs.NewEntity()
	dog.AddHealth(45)
	dog.AddPosition(components.Position{45, 120})

	if dog.HasVelocity() {
		t.Fatal("incorrect component")
	}
	if !dog.HasHealth() {
		t.Fatal("missing component")
	}
	if !dog.HasPosition() {
		t.Fatal("missing component")
	}

	found := false
	ecs.Select(func(e ecs.Entity, health *components.Health) {
		found = true
	})
	if !found {
		t.Fatal("didn't find in select")
	}

	dog.Kill()

	newDog := ecs.NewEntity()
	newDog.AddHealth(45)

	if dog.HasHealth() {
		t.Fatal("dead entity shouldn't have component")
	}
	if !newDog.HasHealth() {
		t.Fatal("new entity should have component")
	}
}

func test1(entity ecs.Entity, health *components.Health) {

}

func test2(entity ecs.Entity, pos *components.Position) {

}
