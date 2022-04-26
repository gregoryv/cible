package cible

import (
	"context"
	"fmt"
	"strings"

	"github.com/gregoryv/logger"
)

func (me *Game) Run(ctx context.Context) error {
gameLoop:
	for {
		select {
		case <-ctx.Done(): // ie. interrupted from the outside
			break gameLoop

		case e := <-me.events: // blocks
			switch e {
			case EventStopGame:
				me.Log(e.Event())
				break gameLoop
			default:
				if err := me.handleEvent(e); err != nil {
					me.Logf("%s: %v", e.Event(), err)
				} else {
					me.Log(e.Event())
				}
			}
		}
	}

	return nil
}

func (me *Game) handleEvent(e Event) error {
	switch e := e.(type) {
	case *EventJoin:
		p := Position{
			Area: "a1", Tile: "01",
		}
		me.Characters = append(me.Characters, Character{
			Ident:    Ident(e.Player.Name),
			Position: p,
		})
		e.joined <- Ident(e.Player.Name)
		return nil

	case *Movement:
		c, err := me.Character(e.Ident)
		// always send a position as someone might be waiting for a
		// response
		defer func() {
			if c == nil {
				return
			}
			e.NewPosition <- c.Position
		}()
		if err != nil {
			return err
		}

		_, t, err := me.Place(c.Position)
		if err != nil {
			return err
		}
		next, err := t.Link(e.Direction)
		if err != nil {
			return err
		}
		if next != "" {
			c.Position.Tile = next

		}
	}
	return nil
}

func Trigger[T Event](g *Game, t T) T {
	g.Events <- t
	return t
}

// ----------------------------------------

const (
	EventStopGame EventString = "stop game"
)

type EventString string

func (me EventString) Event() string { return string(me) }

func MoveCharacter(id Ident, d Direction) *Movement {
	return &Movement{
		Ident:       id,
		Direction:   d,
		NewPosition: make(chan Position, 1),
	}
}

type Movement struct {
	Ident
	Direction

	NewPosition chan Position
}

func (me *Movement) Event() string {
	return fmt.Sprintf("%s move %s", me.Ident, me.Direction)
}

func Join(p Player) *EventJoin {
	return &EventJoin{
		Player: p,
		joined: make(chan Ident, 1), // buffer so event loop doesn't block
		failed: make(chan error, 1),
	}
}

type EventJoin struct {
	Player

	joined chan Ident
	failed chan error
}

func (me *EventJoin) Done() (id Ident, err error) {
	defer me.Close()
	select {
	case id = <-me.joined:
	case err = <-me.failed:
	}
	return
}

func (me *EventJoin) Close() {
	close(me.joined)
	close(me.failed)
}

func (me *EventJoin) Event() string {
	return fmt.Sprintf("%s join", me.Player.Name)
}

type Events chan<- Event

type Event interface {
	Event() string
}

// ----------------------------------------

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
			return &c, nil
		}
	}
	return nil, fmt.Errorf("character %q not found", id)
}

// ----------------------------------------

type Characters []Character

type Character struct {
	Ident
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

func (me *Tile) Link(d Direction) (Ident, error) {
	if d < 0 || int(d) > len(me.Nav) {
		return "", fmt.Errorf("bad direction")
	}
	return me.Nav[int(d)], nil
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

// ----------------------------------------

type Direction int

const (
	N Direction = iota
	E
	S
	W
)
