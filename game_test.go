package cible

import (
	"context"
	"strings"
	"testing"
)

func TestGame(t *testing.T) {
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
	t.Run("handles events", func(t *testing.T) {
		g := NewGame()
		go g.Run(context.Background())
		g.EventChan() <- EventPing
		g.EventChan() <- &EventMove{Direction: E}
	})
	t.Run("handles unknown events", func(t *testing.T) {
		g := NewGame()
		go g.Run(context.Background())
		g.EventChan() <- &EventMove{Direction: Direction(-1)}
	})
}

// gomerge src: world_test.go

func Test_cave(t *testing.T) {
	area := myCave()
	for _, tile := range area.Tiles {
		t.Log(tile, tile.Nav)
	}
}

func myCave() *Area {
	t1 := &Tile{
		Ident: "01",
		Short: "Cave entrance",
		Nav:   Nav{N: "02"},

		Long: `Hidden behind bushes the opening is barely visible.`,
		//
	}

	t2 := &Tile{
		Ident: "02",
		Short: "Fire room",
		Nav:   Nav{E: "03", S: "01"},

		Long: `A small streek of light comes in from a hole in the
		ceiling. The entrance is a dark patch on the west wall, dryer
		than the other walls.`,
		//
	}

	t3 := &Tile{
		Ident: "03",
		Short: "Small area",
		Nav:   Nav{W: "02"},
	}

	return &Area{
		Title: "Cave of Indy",
		Tiles: Tiles{t1, t2, t3},
	}
}

// gomerge src: event_test.go

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
