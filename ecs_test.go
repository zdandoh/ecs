package main

import (
	"fmt"
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
		e.AddPosition(ecs.Position{3, 4})
	}
}

func BenchmarkSelectMatch(b *testing.B) {
	ecs.Reset()

	for i := 0; i < 10000; i++ {
		dog := ecs.NewEntity()
		dog.AddPosition(ecs.Position{100, 100})
		dog.AddHealth(456)
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		c := 0
		ecs.Select(func(entity ecs.Entity, health *ecs.Health, pos *ecs.Position) {
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
		dog.AddPosition(ecs.Position{100, 100})
		dog.AddHealth(456)
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		c := 0
		ecs.Select(func(entity ecs.Entity, v *ecs.Velocity) {
			c++
		})
		if c > 0 {
			b.Fatal("too many entities found")
		}
	}
}

func TestHasComponent(t *testing.T) {
	dog := ecs.NewEntity()
	dog.AddHealth(45)
	dog.AddPosition(ecs.Position{45, 120})

	if dog.HasVelocity() {
		t.Fatal("incorrect component")
	}
	if !dog.HasHealth() {
		t.Fatal("missing component")
	}
	if !dog.HasPosition() {
		t.Fatal("missing component")
	}

	ecs.Select(func(e ecs.Entity, health *ecs.Health) {
		fmt.Printf("Health: %d\n", *health)
	})
}

func test1(entity ecs.Entity, health *ecs.Health) {

}

func test2(entity ecs.Entity, pos *ecs.Position) {

}
