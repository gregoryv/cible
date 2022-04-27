package cible

import (
	"context"
	"errors"
	"fmt"

	"github.com/gregoryv/logger"
)

func NewGame() *Game {
	max := 10
	ch := make(chan Event, max)
	return &Game{
		World:  Earth(),
		Events: ch,
		Logger: logger.Silent,
		Characters: Characters{
			{
				Ident: "god",
				IsBot: true,
			},
		},
		events: ch,
	}
}

type Game struct {
	World
	Characters
	Events chan<- Event // todo maybe always use Trigger?
	logger.Logger

	events chan Event // for reading
}

func (g *Game) Run(ctx context.Context) error {
	g.Log("start game")
eventLoop:
	for {
		select {
		case <-ctx.Done(): // ie. interrupted from the outside
			break eventLoop

		case e := <-g.events: // blocks
			// One event affects the game
			if err := e.Affect(g); err != nil {
				if errors.Is(endEventLoop, err) {
					break eventLoop
				}
				g.Log("event: ", err)
			}
			// Make sure any event can be cleaned up. Triggering
			// side will most likely also wait for event to be
			// done, but this is here to give them the option to
			// ignore it. This does impact performance quite a bit
			// though.
			e.Done()
		}
	}
	close(g.events)
	g.Log("game stopped")
	return nil
}

// Place returns the position as area and tile.
func (g *Game) Place(p Position) (a *Area, t *Tile, err error) {
	if a, err = g.Area(p.Area); err != nil {
		return
	}
	t, err = a.Tile(p.Tile)
	return
}

// Character returns a character in the game by id.
func (g *Game) Character(id Ident) (*Character, error) {
	for _, c := range g.Characters {
		if c.Ident == id {
			return c, nil
		}
	}
	return nil, fmt.Errorf("character %q not found", id)
}

// ----------------------------------------

type Characters []*Character

type Character struct {
	Ident
	Name
	Position
	IsBot
}

type IsBot bool

type Player struct {
	Name
}

type Bot struct{}

type Position struct {
	Area Ident
	Tile Ident
}

func (p *Position) Equal(v Position) bool {
	return p.Area == v.Area && p.Tile == v.Tile
}

type Ident string
type Name string
type Short string
type Long string
type Title string
