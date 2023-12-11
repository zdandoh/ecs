# ecs
Proof of concept Golang ECS library using `go generate`

### About
This library is a code generator that creates a bespoke entity component system library based off the provided input. The goal of this library is to provide
a highly ergonomic, performant, and type safe ECS API in Go without using reflection. The cost is that you must frequently regenerate the emitted library using `go generate`.
The code generator supports an unlimited number of components, and can grow its entity pool during runtime.
The code generator requires two inputs:
1. A file containing only component definitions
1. A root package that contains code that uses the generated ECS library

# How to Use
1. Create a file containing all your component definitions. This file will be copied into the generated ECS library
```go
package main // just an example package - could be anything

type Position struct {
	X int
	Y int
}

type Name string

type Health int
```

1. Add a `go generate` directive to one of your source files.
```go
package main

// This will generate a package named "myecspkg" when go generate is run.
//go:generate go run github.com/zdandoh/ecs/codegen myecspkg /path/to/components.go /path/to/this/package

import (
	"fmt"
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
    
    dog.AddPosition(ecs.Position{45, 120})
    
    // Run efficient ECS queries without reflection
    ecs.Select(func(entity ecs.Entity, name *ecs.Name, hp *ecs.Health) {
        fmt.Println("%s has %d health", *name, *hp)
    })
    ecs.Select(func(entity ecs.Entity, hp *ecs.Health) {
        *hp -= 1
        if hp <= 0 {
            entity.Kill()
            if entity.Name() != nil {
                fmt.Println("%s died!", entity.Name())
            }
        }
    })
}
```
If you use this library you probably will want to run `go generate` as a
pre-build step.

### How It Works
The code generator uses the provided component definitions to generate
helper functions and storage data structures for each component, but also
needs to know which package is going to consume the generated package! Kinda
weird right? This is necessary because the library needs to know which subsets
of components might be queried against so that it can generate code to serve
those queries. This allows the generated library to fully avoid reflection
while maintaining the nice selection syntax. This probably isn't a great idea,
but it seems cool!