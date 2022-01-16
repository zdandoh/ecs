package ecs

import (
	"fmt"
	"testing"
)

func TestRange(t *testing.T) {
	r := componentRange{}
	r.Add(45)
	fmt.Println(r)
}
