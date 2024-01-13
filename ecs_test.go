package main

import (
	"fmt"
	"testing"
	"time"

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

func BenchmarkIntMap(b *testing.B) {
	m := make(map[ecs.EntityID]uint64)
	for n := 0; n < b.N; n++ {
		m[ecs.EntityID(n/10)] = 45
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

func TestDefault(t *testing.T) {
	ecs.Reset()

	dog := ecs.NewEntity()
	dog.SetHealth(components.Health(45))
	hp := dog.DefaultHealth(components.Health(65))
	v := dog.DefaultVelocity(components.Velocity{20, 20})
	if *hp != 45 || v.X != 20 || v.Y != 20 {
		t.Fatal(hp, v)
	}
}

func TestIsEntity(t *testing.T) {
	ecs.Reset()

	dog := ecs.NewEntity()
	dog.SetHealth(45)

	var other ecs.Entity
	ecs.Select(func(e ecs.Entity, hp *components.Health) {
		other = e
	})

	if dog == other {
		t.Fatal()
	}
	if !dog.Is(other) {
		t.Fatal()
	}
}

func TestSelectSorted(t *testing.T) {
	ecs.Reset()

	e1 := ecs.NewEntity()
	e2 := ecs.NewEntity()
	e3 := ecs.NewEntity()
	e4 := ecs.NewEntity()

	e1.SetHealth(67)
	e2.SetHealth(43)
	e4.SetHealth(100)
	e3.SetPosition(components.Position{3, 3})

	ecs.SelectSorted(func(a ecs.Entity, b ecs.Entity) int {
		return int(*a.Health() - *b.Health())
	}, func(e ecs.Entity, hp *components.Health) {
		fmt.Println(*hp)
	})
}

func BenchmarkEntitySelect(b *testing.B) {
	ecs.Reset()

	for i := 0; i < 10000; i++ {
		e := ecs.NewEntity()
		e.SetPos(components.Pos{
			X: 45,
			Y: 3846,
		})
		e.SetVel(components.Vel{
			X: 38456,
			Y: 1234,
		})
	}
	b.ResetTimer()

	for n := 0; n < b.N; n++ {
		ecs.Select(func(e ecs.Entity, pos *components.Pos, vel *components.Vel) {
			pos.X += vel.X
			pos.Y += vel.Y
		})
	}
}

func BenchmarkEntityCreation2(b *testing.B) {
	ecs.Reset()

	s := time.Now()
	for i := 0; i < 1000; i++ {
		e := ecs.NewEntity()
		e.SetPos(components.Pos{
			X: 45,
			Y: 3846,
		})
		e.SetVel(components.Vel{
			X: 38456,
			Y: 1234,
		})
	}
	for i := 0; i < 9000; i++ {
		e := ecs.NewEntity()
		e.SetPos(components.Pos{
			X: 45,
			Y: 3846,
		})
	}
	fmt.Println(time.Since(s))

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		ecs.Select(func(e ecs.Entity, vel *components.Vel, pos *components.Pos) {
			pos.X += vel.X
			pos.Y += vel.Y
		})
	}
}

func BenchmarkRemoveComponent(b *testing.B) {
	ecs.Reset()

	s := time.Now()
	for i := 0; i < 1000; i++ {
		e := ecs.NewEntity()
		e.SetPos(components.Pos{
			X: 45,
			Y: 4567,
		})
	}
	fmt.Println(time.Since(s))

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		ecs.Select(func(e ecs.Entity, pos *components.Pos) {
			e.SetVel(components.Vel{
				X: 12345,
				Y: 67899,
			})
		})
		ecs.Select(func(e ecs.Entity, pos *components.Pos, vel *components.Vel) {
			e.RemoveVel()
		})
	}
}

func TestComponentID(t *testing.T) {
	ecs.Reset()

	if ecs.VelocityID == ecs.PosID {
		t.Fatal()
	}
	if ecs.VelocityID != ecs.VelocityID {
		t.Fatal()
	}
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

func TestSelectDead(t *testing.T) {
	ecs.Reset()

	e1 := ecs.NewEntity()
	e1.SetPos(components.Pos{})
	e1.Kill()

	c := 0
	ecs.Select(func(e ecs.Entity, p *components.Pos) {
		c++
	})
	if c > 0 {
		t.Fatal(c)
	}
}

func TestRelationship(t *testing.T) {
	ecs.Reset()

	e := ecs.NewEntity()
	apple := ecs.NewEntity()
	gun := ecs.NewEntity()
	egg := ecs.NewEntity()

	e.SetHas(apple, components.Has{Count: 5})
	e.SetHas(gun, components.Has{Count: 1})
	e.SetHas(egg, components.Has{Count: 3})

	e.SetLikes(apple)
	e.SetLikes(gun)
	e.SetLikes(egg)

	egg.Kill()

	count := 0
	e.EachHas(func(e ecs.Entity, has *components.Has) {
		count += has.Count
	})
	if count != 6 {
		t.Fatal(count)
	}

	count = 0
	e.EachLikes(func(e ecs.Entity) {
		count++
	})
	if count != 2 {
		t.Fatal(count)
	}

	if e.Likes(egg) {
		t.Fatal()
	}
	if !e.Likes(gun) {
		t.Fatal()
	}

	e.RemoveLikes(gun)
	if e.Likes(gun) {
		t.Fatal()
	}

	single := ecs.NewEntity()
	single.Likes(gun)
	single.RemoveLikes(gun)
	if single.AnyLikes() {
		t.Fatal()
	}
}

func BenchmarkRelationshipBuild(b *testing.B) {
	ecs.Reset()

	e := ecs.NewEntity()
	ents := make([]ecs.Entity, 0)
	for i := 0; i < 1000; i++ {
		n := ecs.NewEntity()
		ents = append(ents, n)
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		for _, ent := range ents {
			e.SetLikes(ent)
		}
	}
}

func BenchmarkRelationshipIter(b *testing.B) {
	ecs.Reset()

	e := ecs.NewEntity()
	for i := 0; i < 1000; i++ {
		n := ecs.NewEntity()
		e.SetLikes(n)
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		e.EachLikes(func(e ecs.Entity) {
		})
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
