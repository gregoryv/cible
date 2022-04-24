package cible

import (
	"context"
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
			me.handleEvent(e)
		}
	}

	return nil
}

func (me *Game) handleEvent(e Event) {
	// todo handle event
}

// EventChan returns a channel for adding events to the game
func (me *Game) EventChan() chan<- Event {
	return me.events
}

type Event interface {
	Event() string
}
