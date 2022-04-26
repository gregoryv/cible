package cible

import (
	"context"
	"strings"
	"testing"
)

func TestGame_play(t *testing.T) {
	g := NewGame()
	g.Logger = t
	ctx, cancel := context.WithCancel(context.Background())
	go g.Run(ctx)
	defer cancel()

	p := Player{Name: "John"}

	cid, err := Trigger(g, Join(p)).Done() // blocks
	if err != nil {
		t.Fatal(err)
	}

	g.Events <- MoveCharacter(cid, W)   // nothing ther)
	g.Events <- MoveCharacter("Eve", N) // no such playe)
	g.Events <- MoveCharacter("god", N) // cannot be move)
	g.Events <- MoveCharacter(cid, Direction(-1))
	g.Events <- &badEvent{}

	pos := <-Trigger(g, MoveCharacter(cid, N)).NewPosition
	t.Log(pos)

	pos = <-Trigger(g, MoveCharacter(cid, E)).NewPosition
	t.Log(pos)

	pos = <-Trigger(g, MoveCharacter(cid, W)).NewPosition
	t.Log(pos)
	if pos.Tile != "02" {
		t.Error("got", pos.Tile, "exp", "02")
	}
}

func TestEventStopGame(t *testing.T) {
	g := NewGame()
	go g.Run(context.Background())
	g.Events <- EventStopGame
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
