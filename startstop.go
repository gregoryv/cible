package cible

import (
	"fmt"
)

func StopGame() *EventStopGame {
	return &EventStopGame{}
}

type EventStopGame struct{}

func (e *EventStopGame) Affect(g *Game) error {
	// special event that ends the loop, thus we do things here as
	// no other events should be affecting the game
	g.Log("shutting down...")

	return endEventLoop
}

var endEventLoop = fmt.Errorf("end event loop")
