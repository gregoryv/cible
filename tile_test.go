package cible

import (
	"fmt"
	"testing"
)

func ExampleTile_Link() {
	t1 := &Tile{Ident: "t1"}
	t2 := &Tile{Ident: "t2"}
	t1.Link(t2, N)
	t2.Link(t1, S) // already linked by first
	fmt.Println(t1.Nav.String())
	// output:
	// N:t2
}

func TestTile_Link(t *testing.T) {
	defer func() {
		e := recover()
		if e == nil {
			t.Fatal("expect panic if trying to override link")
		}
	}()
	t1 := &Tile{Ident: "t1"}
	t2 := &Tile{Ident: "t2"}
	t3 := &Tile{Ident: "t3"}
	t1.Link(t2, N) // first is ok
	t1.Link(t3, N) // but you should not be able to override it
}
