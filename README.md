# ecs
Proof of concept Golang ECS library using `go generate`

# How to Use
This library uses a file of component definitions to generate an ECS package for those components. This approach allows for good ergonomics and efficiency. In order to use this library, create a file that contains type definitions for components (and any associated helper methods for those components) and run the code generator provided by this library. Below is an example:
```go
//go:generate go run github.com/zdandoh/ecs/codegen ecs 10000000

type Position struct {
	X int
	Y int
}

func (p Position) Dist(p2 Position) float64 {
	return math.Sqrt(math.Pow(float64(p.X - p2.X), 2) + math.Pow(float64(p.Y - p2.Y), 2))
}

type Name string

type Health int
```

To build, simply run `go generate <components_file>.go`. The generated `ecs` package can then be used for efficient ecs operations:

```go
// Create a few new entities
cat := ecs.NewEntity()
dog := ecs.NewEntity()

// Give them some components
cat.AddHealth(46)
dog.AddHealth(100)

cat.AddName("mixer")
dog.AddName("rex")

dog.AddPosition(ecs.Position{45, 120})

// Use the APIs provided by the package for efficient querying
ecs.SelectWithComponent(func(e ecs.Entity) {
	if *e.Name() == "rex" {
		// Since a pointer to the component is passed to this callback, we can mutate the value
		*e.Health() -= 1
	}
}, ecs.ComponentName)

ecs.SelectWithComponents(func(e ecs.Entity) {
	fmt.Printf("%s has %d health\n", *e.Name(), *e.Health())
}, ecs.ComponentHealth, ecs.ComponentName)

dog.Kill()

ecs.SelectWithComponents(func(e ecs.Entity) {
	// The "dog" entity is no longer returned in queries
	fmt.Printf("%s has %d health\n", *e.Name(), *e.Health())
}, ecs.ComponentHealth, ecs.ComponentName)
```

### About
This library is far from complete, and is mostly just a first pass at what a performant ecs library with a nice API might look like in Go. The implementation allocates minimally at runtime and tries to be cache efficient. Creating new entities or components does not allocate. Queries should execute very quickly, even with tens of thousands of entities. Below are some areas that need additional attention:
- Cleaner generated code
- A limit of 64 components are supported for a generated package
- A hardcoded maximum number of entities is supported. This number can be specified at package generation time
