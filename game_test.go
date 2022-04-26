package cible

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestGame(t *testing.T) {
	g := NewGame()
	ctx, cancel := context.WithCancel(context.Background())
	go g.Run(ctx)
	defer cancel()

	p := Player{Name: "John"}

	t.Run("handles events", func(t *testing.T) {
		g.Events <- &Join{Player: p}
		g.Events <- &Move{Player: p, Direction: E}
	})

	t.Run("handles unknown events", func(t *testing.T) {
		g.Events <- &Move{Direction: Direction(-1)}
	})

	// let all events pass
	<-time.After(10 * time.Millisecond)
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

func TestMove(t *testing.T) {
	e := &Move{Player{Name: "John"}, S}
	if got := e.Event(); !strings.Contains(got, "John") {
		t.Errorf("missing name: %q", got)
	}
}
