package cible

import (
	"context"
	"strings"
	"testing"
)

func TestGame(t *testing.T) {
	g := NewGame()
	ctx, cancel := context.WithCancel(context.Background())
	go g.Run(ctx)
	defer cancel()

	p := Player{Name: "John"}

	t.Run("handles events", func(t *testing.T) {
		g.Events <- &EventJoin{Player: p}
		g.Events <- &EventMove{Player: p, Direction: E}
	})

	t.Run("handles unknown events", func(t *testing.T) {
		g.Events <- &EventMove{Direction: Direction(-1)}
	})
}

func TestEventStopGame(t *testing.T) {
	g := NewGame()
	go g.Run(context.Background())
	g.Events <- EventStopGame
}

func Test_cave(t *testing.T) {
	area := myCave()
	for _, tile := range area.Tiles {
		t.Log(tile, tile.Nav)
	}
}

func TestEventMove(t *testing.T) {
	p := Player{
		Name: "John",
	}
	e := &EventMove{p, S}
	got := e.Event()
	if !strings.Contains(got, "John") {
		t.Errorf("missing name: %q", got)
	}
}
