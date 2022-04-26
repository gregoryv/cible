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

	_ = Trigger(g, &Join{Player: p})
	// currently the name is used as an identifier of characters

	g.Events <- MoveCharacter("John", W) // nothing ther)
	g.Events <- MoveCharacter("Eve", N)  // no such playe)
	g.Events <- MoveCharacter("god", N)  // cannot be move)
	g.Events <- MoveCharacter("John", Direction(-1))
	g.Events <- &badEvent{}

	e := Trigger(g, MoveCharacter("John", N))

	t.Log(<-e.NewPosition)
	//t.Fail()
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
