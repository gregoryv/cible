package cible

import (
	"context"
	"fmt"
	"strings"

	"github.com/gregoryv/logger"
)

func NewGame() *Game {
	max := 10
	ch := make(chan Event, max)
	return &Game{
		World:  Earth(),
		Events: ch,
		Logger: logger.Silent,

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

func (me *Game) Run(ctx context.Context) error {
gameLoop:
	for {
		select {
		case <-ctx.Done(): // ie. interrupted from the outside
			break gameLoop

		case e := <-me.events: // blocks
			me.Log(e.Event())
			switch e {
			case EventStopGame:
				break gameLoop
			default:
				me.handleEvent(e)
			}
		}
	}

	return nil
}

func (me *Game) handleEvent(e Event) {
	me.Log(e)
	switch e := e.(type) {
	case *Join:
		me.Characters = append(me.Characters, Character{
			Name: e.Player.Name,
		})
	case *MoveCharacter:

	}
}

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
		Title: "Cave of Indy",
		Tiles: Tiles{t1, t2, t3},
	}
}

// ----------------------------------------

type Characters []Character

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

type Areas []*Area

type Area struct {
	Ident
	Title
	Tiles
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
	Area int
	Tile int
}

type Ident string

func (me Ident) Id() string { return string(me) }

type Name string
type Short string
type Long string
type Title string

// ----------------------------------------

const (
	EventStopGame EventString = "stop game"
	EventPing     EventString = "ping"
)

type EventString string

func (me EventString) Event() string { return string(me) }

type MoveCharacter struct {
	Name
	Direction
}

func (me *MoveCharacter) Event() string {
	return fmt.Sprintf("%s moves %s", me.Name, me.Direction)
}

type Join struct {
	Player
}

func (me *Join) Event() string {
	return fmt.Sprintf("%s joined game", me.Player.Name)
}

type Events chan<- Event

type Event interface {
	Event() string
}

// ----------------------------------------

type Direction int

const (
	N Direction = iota
	E
	S
	W
)
