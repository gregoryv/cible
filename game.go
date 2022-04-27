package cible

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gregoryv/logger"
)

func (me *Game) Run(ctx context.Context) error {
	me.Log("start game")
eventLoop:
	for {
		select {
		case <-ctx.Done(): // ie. interrupted from the outside
			break eventLoop

		case e := <-me.events: // blocks
			// One event affects the game
			if err := e.Affect(me); err != nil {
				if errors.Is(endEventLoop, err) {
					break eventLoop
				}
				me.Log("event: ", err)
				continue
			}
			// Make sure any event can be cleaned up. Triggering
			// side will most likely also wait for event to be
			// done, but this is here to give them the option to
			// ignore it. This does impact performance quite a bit
			// though.
			go e.Done()
		}
	}
	me.Log("game stopped")
	return nil
}

func Trigger[T Event](g *Game, t T) T {
	g.Events <- t
	return t
}

type Events chan<- Event

type Event interface {
	Affect(*Game) error // called in the event loop

	// Done blocks until event is handled, can be called multiple
	// times.
	Done() error
}

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
	Events
	logger.Logger

	events chan Event
}

func (me *Game) Place(p Position) (a *Area, t *Tile, err error) {
	if a, err = me.Area(p.Area); err != nil {
		return
	}
	t, err = a.Tile(p.Tile)
	return
}

func (me *Game) Character(id Ident) (*Character, error) {
	for _, c := range me.Characters {
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

type World struct {
	Areas
}

func (me *World) Area(id Ident) (*Area, error) {
	for _, a := range me.Areas {
		if a.Ident == id {
			return a, nil
		}
	}
	return nil, fmt.Errorf("area %q not found", id)
}

type Areas []*Area

type Area struct {
	Ident
	Title
	Tiles
}

func (me *Area) Tile(id Ident) (*Tile, error) {
	for _, t := range me.Tiles {
		if t.Ident == id {
			return t, nil
		}
	}
	return nil, fmt.Errorf("tile %q not found", id)
}

type Tiles []*Tile

type Tile struct {
	Ident
	Short
	Long
	Nav
}

func (me *Tile) String() string {
	return fmt.Sprintf("%s %s", me.Ident, me.Short)
}

type Nav [4]Ident

func (me Nav) String() string {
	var res []string
	for d, id := range me {
		if id != "" {
			res = append(res, Direction(d).String()+":"+string(id))
		}
	}
	return strings.Join(res, " ")
}

type Position struct {
	Area Ident
	Tile Ident
}

func (me *Position) Equal(p Position) bool {
	return me.Area == p.Area && me.Tile == p.Tile
}

type Ident string
type Name string
type Short string
type Long string
type Title string

// ----------------------------------------

func Earth() World {
	return World{
		Areas: Areas{myCave()},
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
		Ident: "a1",
		Title: "Cave of Indy",
		Tiles: Tiles{t1, t2, t3},
	}
}
