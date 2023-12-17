# ecs
Proof of concept Golang ECS library using `go generate`

### About
This library is a code generator that creates a bespoke entity component system library based off the provided input. The goal of this library is to provide
a highly ergonomic, performant, and type safe ECS API in Go without using reflection. The cost is that you must frequently regenerate the emitted library using `go generate`.
The code generator supports an unlimited number of components, and can grow its entity pool during runtime.
The code generator takes a package containing only component definitions as an input.

# How to Use
1. Create a "root" package that will use the generated ECS package. The root package must
have go modules setup.
```go
package main

import "fmt"

func main() {
	fmt.Println("My new package!")
}
```

2. Create a subpackage within the root package containing all your component definitions.
```go
package components

type Position struct {
	X int
	Y int
}

type Name string

type Health int
```

3. Add a `go generate` directive to one of the source files of your root package. This
directive takes the name of the generated package and the name of the component package as
arguments.
```go
package main

// This will generate a package named "myecspkg" when go generate is run.
//go:generate go run github.com/zdandoh/ecs/codegen myecspkg components

import "fmt"

func main() {
	fmt.Println("My new package!")
}
```

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
    cat.AddHealth(46)
    dog.AddHealth(100)
    
    cat.AddName("mixer")
    dog.AddName("rex")
    
    dog.AddPosition(components.Position{45, 120})
    
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

### How It Works
The code generator uses the provided component definitions to generate
helper functions and storage data structures for each component, but also
analyzes the root package to determine which component queries are made.
This is necessary because the library needs to know which subsets
of components might be queried against so that it can generate code to serve
those queries. This allows the generated library to fully avoid reflection
while maintaining the nice selection syntax. This probably isn't a great idea,
but it seems cool!