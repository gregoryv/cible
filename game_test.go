package cible

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestGame_play(t *testing.T) {
	g := NewGame()
	g.Logger = t
	ctx, cancel := context.WithCancel(context.Background())
	go g.Run(ctx)
	defer cancel()

	p := Player{Name: "John"}

	g.Events <- &Join{Player: p}
	// currently the name is used as an identifier of characters
	g.Events <- &MoveCharacter{"John", N}
	g.Events <- &MoveCharacter{"Eve", N} // no such player
	g.Events <- &MoveCharacter{"god", N} // cannot be moved

	g.Events <- &MoveCharacter{"John", Direction(-1)}

	// let all events pass
	<-time.After(10 * time.Millisecond)
	//	t.Fail()
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

func TestMoveCharacter(t *testing.T) {
	e := &MoveCharacter{"John", S}
	if got := e.Event(); !strings.Contains(got, "John") {
		t.Errorf("missing name: %q", got)
	}
}
