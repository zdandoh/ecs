package main

import (
	"testing"

	"github.com/zdandoh/ecs/components"
	ecs "github.com/zdandoh/ecs/ecspkg"
	"github.com/zdandoh/ecs/ecspkg/entity"
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
		e.SetPosition(components.Position{3, 4})
	}
}

func BenchmarkSelectMatch(b *testing.B) {
	ecs.Reset()

	for i := 0; i < 10000; i++ {
		dog := ecs.NewEntity()
		dog.SetPosition(components.Position{100, 100})
		dog.SetHealth(456)
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
		dog.SetPosition(components.Position{100, 100})
		dog.SetHealth(456)
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

func TestValidEmptyRef(t *testing.T) {
	ecs.Reset()

	dog := ecs.NewEntity()
	var ref entity.Ref
	if ecs.Lookup(ref).Alive() {
		t.Fatal()
	}
	_ = dog
}

func TestStopEarly(t *testing.T) {
	ecs.Reset()

	dog := ecs.NewEntity()
	cat := ecs.NewEntity()

	dog.SetHealth(54)
	cat.SetHealth(45)

	count := 0
	ecs.Select(func(e ecs.Entity, health *components.Health) bool {
		count++
		return false
	})
	if count != 1 {
		t.Fatal(count)
	}
}

func TestGetAllComponents(t *testing.T) {
	ecs.Reset()

	dog := ecs.NewEntity()
	dog.SetHealth(45)
	dog.SetPosition(components.Position{1, 1})
	dog.SetVelocity(components.Velocity{2, 2})
	if len(dog.Components()) != 3 {
		t.Fatal(dog.Components())
	}
}

func TestHasComponent(t *testing.T) {
	ecs.Reset()

	dog := ecs.NewEntity()
	dog.SetHealth(45)
	dog.SetPosition(components.Position{45, 120})

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
	newDog.SetHealth(45)

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
