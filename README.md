# ecs
Golang ECS library using `go generate`

## About
This library is a code generator that creates a bespoke entity component system library based off the provided input. The goal of this library is to provide
a highly ergonomic, performant, and type safe ECS API in Go without using reflection. The cost is that you must frequently regenerate the emitted library using `go generate`.
The code generator supports an unlimited number of components, and can grow its entity pool during runtime.
The code generator takes a package containing only component definitions as an input.

## Example
The following component definitions:
```go
type Position struct {
	X int
	Y int
}

type Name string

type Health int
```
Can be used like this!
```go
func main() {
    // Create some entities
    cat := ecs.NewEntity()
    dog := ecs.NewEntity()
    
    // Give them some components
    cat.SetHealth(46)
    dog.SetHealth(100)
    
    cat.SetName("mixer")
    dog.SetName("rex")
    
    dog.SetPosition(components.Position{45, 120})
    
    // Run efficient ECS queries without reflection. Pointers to the components
    // are passed to Select, allowing for direct reading and updating of attributes
    ecs.Select(func(entity ecs.Entity, name *components.Name, hp *components.Health) {
        fmt.Println("%s has %d health", *name, *hp)
    })
    ecs.Select(func(entity ecs.Entity, hp *components.Health) {
        *hp -= 1
        if hp <= 0 {
            if entity.HasName() {
                fmt.Printf("%s died!\n", entity.Name())
            }
            entity.Kill()
        }
    })
}
```

This library has basic support for relationships, which are special
components that allow for the modeling of one-to-many relationships in an ecs.
To define a new relationship, annotate a component with a `Relationship` field:
```go
type Has struct {
    Relationship struct{}
    Count        int
}
```
The code generator generates new helpers that are specific to relationships:
```go
boy := ecs.NewEntity()
dog := ecs.NewEntity()
apple := ecs.NewEntity()

boy.SetHas(dog, components.Has{Count: 1})
boy.SetHas(apple, components.Has{Count: 5})

count := 0
e.EachHas(func(e ecs.Entity, has *components.Has) {
    count += has.Count
})
fmt.Println("The boy has %d things", count)
```

## How to Use
1. Create (or use a pre-existing) Go module that will use the generated ECS package. For this example, assume the
following structure:
```
mygomodule/
├─ go.mod
├─ components/
│  ├─ components.go
├─ main.go
```

2. Run `go get -tool github.com/zdandoh/ecs/codegen`

3. Create a package within your module containing all your component definitions. This package should contain only component definitions.
```go
package components

type Position struct {
	X int
	Y int
}

type Name string

type Health int
```

4. Add a `go generate` directive to any source file in your module. This
directive takes the name of the input component package and the name of the 
package to be generated as arguments.
```go
package main

// This will generate a package named "myecspkg" when go generate is run.
//go:generate go run github.com/zdandoh/ecs/codegen myecspkg components

import "fmt"

func main() {
	fmt.Println("My new package!")
}
```

5. Run `go generate`. The tool will automagically scan your module for component
queries that it needs to generate code for.

You're done! The generated ECS package can be imported and used
```go
package main

// This will generate a package named "myecspkg" when go generate is run.
//go:generate go run github.com/zdandoh/ecs/codegen myecspkg components

import (
	"fmt"
	"components"
	ecs "myecspkg"
)

func main() {
    cat := ecs.NewEntity()
    dog := ecs.NewEntity()
    
    // Give them some components
    cat.SetHealth(46)
    dog.SetHealth(100)
    
    cat.SetName("mixer")
    dog.SetName("rex")
    
    dog.SetPosition(components.Position{45, 120})
    
    // Run efficient ECS queries without reflection
    ecs.Select(func(entity ecs.Entity, name *components.Name, hp *components.Health) {
        fmt.Println("%s has %d health", *name, *hp)
    })
    ecs.Select(func(entity ecs.Entity, hp *components.Health) {
        *hp -= 1
        if hp <= 0 {
            if entity.HasName() {
                fmt.Printf("%s died!\n", entity.Name())
            }
            entity.Kill()
        }
    })
}
```
If you use this library you probably will want to run `go generate` as a
pre-build step.

## Performance
Performance is a major concern for an ECS library. This library implements a bitset ECS,
so average query performance will be slower than an archetype based ECS
like [arche](https://github.com/mlange-42/arche) but with significantly faster component addition and removal times.

The generated code is efficient. On my machine, a query like:
```go
ecs.Select(func(e ecs.Entity, pos *components.Pos, vel *components.Vel) {
    pos.X += vel.X
    pos.Y += vel.Y
})
```
Costs 3.75 ns per matching entity, and 0.75 ns per non-matching entity. 

## How It Works
The code generator uses the provided component definitions to generate
helper functions and storage data structures for each component, but also
analyzes the module to determine which component queries are made.
This is necessary because the library needs to know which subsets
of components might be queried against so that it can generate code to serve
those queries. This allows the generated library to fully avoid reflection
while maintaining the nice selection syntax.
