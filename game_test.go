package cible

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"testing"
)

func TestGame_play(t *testing.T) {
	g := NewGame()
	g.Logger = t
	go g.Run(context.Background())
	defer func() {
		Trigger(g, StopGame()).Done()
	}()

	p := Player{Name: "John"}
	cid, err := Trigger(g, Join(p)).Done()
	if err != nil {
		t.Fatal(err)
	}

	Trigger(g, MoveCharacter(cid, N)).Done()
	Trigger(g, MoveCharacter(cid, E)).Done()

	pos, _ := Trigger(g, MoveCharacter(cid, W)).Done()
	if pos.Tile != "02" {
		t.Error("got", pos.Tile, "exp", "02")
	}
	_, tile, err := g.Place(pos)
	if err != nil {
		t.Fatal(tile, err)
	}
	if err := Trigger(g, Leave(cid)).Done(); err != nil {
		t.Error(err)
	}
}

func Test_badEvents(t *testing.T) {
	g := NewGame()
	go g.Run(context.Background())
	defer g.Stop()

	p := Player{Name: "John"}
	cid, err := Trigger(g, Join(p)).Done() // blocks
	if err != nil {
		t.Fatal(err)
	}

	g.Events <- MoveCharacter("Eve", N) // no such playe)
	g.Events <- MoveCharacter("god", N) // cannot be move)
	g.Events <- MoveCharacter(cid, Direction(-1))
	g.Events <- &badEvent{}
	_, _ = Trigger(g, MoveCharacter(cid, W)).Done()
	e := Trigger(g, Leave("no such"))
	if err := e.Done(); err == nil {
		t.Error("Leave unknown cid should fail")
	}
}

func TestEvent_Done(t *testing.T) {
	g := NewGame()
	go g.Run(context.Background())
	defer g.Stop()
	t.Run("MoveCharacter", func(t *testing.T) {
		defer catchPanic(t)
		e := Trigger(g, MoveCharacter("x", W))
		e.Done()
		e.Done()
	})

	t.Run("StopGame", func(t *testing.T) {
		defer catchPanic(t)
		e := Trigger(g, StopGame())
		e.Done()
		e.Done()
	})

}

func catchPanic(t *testing.T) {
	if err := recover(); err != nil {
		t.Helper()
		t.Fatal(err)
	}
}

func Test_cancelGame(t *testing.T) {
	g := NewGame()
	ctx, cancel := context.WithCancel(context.Background())
	go g.Run(ctx)
	cancel()
}

func Test_cave(t *testing.T) {
	for _, tile := range myCave().Tiles {
		t.Log(tile, tile.Nav)
	}
}

func TestMovement(t *testing.T) {
	e := MoveCharacter("John", S)
	if got := e.Event(); !strings.Contains(got, "John") {
		t.Errorf("missing name: %q", got)
	}
}

func TestArea_Tile(t *testing.T) {
	var a Area
	if _, err := a.Tile("x"); err == nil {
		t.Fail()
	}
}

type badEvent struct{}

func (me *badEvent) Event() string { return "badEvent" }

// ----------------------------------------

func BenchmarkMoveCharacter_1_player(b *testing.B) {
	g := NewGame()
	go g.Run(context.Background())
	defer g.Stop()

	p := Player{Name: "John"}
	cid, err := Trigger(g, Join(p)).Done() // blocks
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < b.N; i++ {
		Trigger(g, MoveCharacter(cid, N)).Done()
		Trigger(g, MoveCharacter(cid, S)).Done()
	}
}

func BenchmarkMoveCharacter_1000_player(b *testing.B) {
	g := NewGame()
	go g.Run(context.Background())
	defer g.Stop()

	for i := 0; i < 1000; i++ {
		p := Player{Name: Name(fmt.Sprintf("John%v", i))}
		_, err := Trigger(g, Join(p)).Done() // blocks
		if err != nil {
			b.Fatal(err)
		}
	}

	for i := 0; i < b.N; i++ {
		cid := Ident(fmt.Sprintf("John%v", rand.Intn(1000)))
		Trigger(g, MoveCharacter(cid, N)).Done()
		Trigger(g, MoveCharacter(cid, S)).Done()
	}
}
