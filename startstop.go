package cible

import (
	"fmt"
	"sync"
)

func StopGame() *EventStopGame {
	return &EventStopGame{
		failed: make(chan error, 1),
	}
}

type EventStopGame struct {
	err error // set when done

	sync.Once
	failed chan error
}

func (e *EventStopGame) Affect(g *Game) error {
	// special event that ends the loop, thus we do things here as
	// no other events should be affecting the game
	g.Log("shutting down...")

	e.failed <- nil
	return endEventLoop
}

var endEventLoop = fmt.Errorf("end event loop")

func (e *EventStopGame) Done() error {
	e.Once.Do(func() {
		e.err = <-e.failed
		close(e.failed)
	})
	return e.err
}

func (e *EventStopGame) setErr(v error) {
	go e.Done()
	e.failed <- v
}
