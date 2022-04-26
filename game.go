package cible

import (
	"context"
	"fmt"
	"strings"

	"github.com/gregoryv/logger"
)

func NewGame() *Game {
	return &Game{
		events: make(chan Event),
		Logger: logger.New(),
	}
}

type Game struct {
	events chan Event
	logger.Logger
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
	// todo handle event
	me.Log(e)
}

// EventChan returns a channel for adding events to the game
func (me *Game) EventChan() chan<- Event {
	return me.events
}

type Player struct {
	Name
}

type Name string

// gomerge src: world.go

type Area struct {
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

type Ident string
type Short string
type Long string
type Title string

// gomerge src: event.go

type Event interface {
	Event() string
}

const (
	EventStopGame EventString = "stop game"
	EventPing     EventString = "ping"
)

type EventString string

func (me EventString) Event() string { return string(me) }

type EventMove struct {
	Player
	Direction
}

func (me *EventMove) Event() string {
	return fmt.Sprintf("%s moves %s", me.Player.Name, me.Direction)
}

// gomerge src: direction.go

type Direction int

const (
	N Direction = iota
	E
	S
	W
)

func (me Direction) Opposite() Direction {
	switch me {
	case N:
		return S
	case E:
		return W
	case W:
		return E
	}
	return me
}
