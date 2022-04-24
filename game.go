package cible

import (
	"context"
	"log"
)

func NewGame() *Game {
	return &Game{
		events: make(chan Event),
	}
}

type Game struct {
	events chan Event
}

func (me *Game) Run(ctx context.Context) error {

gameLoop:
	for {
		select {
		case <-ctx.Done(): // ie. interrupted from the outside
			break gameLoop

		case e := <-me.events: // blocks
			log.Println(e.Event())
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
	log.Println(e)
}

// EventChan returns a channel for adding events to the game
func (me *Game) EventChan() chan<- Event {
	return me.events
}

type Event interface {
	Event() string
}

const (
	EventStopGame EventString = "stop game"
	EventPing     EventString = "ping"
)

type EventString string

func (me EventString) Event() string { return string(me) }
