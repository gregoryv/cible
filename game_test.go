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

	t.Run("handles events", func(t *testing.T) {
		g.EventChan() <- EventPing
		g.EventChan() <- &EventMove{Direction: E}
	})
	t.Run("handles unknown events", func(t *testing.T) {
		g.EventChan() <- &EventMove{Direction: Direction(-1)}
	})
}

func TestGame_exits(t *testing.T) {
	t.Run("can be cancelled", func(t *testing.T) {
		g := NewGame()
		ctx, cancel := context.WithCancel(context.Background())
		go g.Run(ctx)
		cancel()
	})
	t.Run("can be stopped", func(t *testing.T) {
		g := NewGame()
		go g.Run(context.Background())
		g.EventChan() <- EventStopGame
	})
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
