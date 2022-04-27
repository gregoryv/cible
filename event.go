package cible

import "fmt"

// Trigger an event which affects the game. Callers should call Done
// to wait for the event.
func Trigger[T Event](g *Game, t T) (r T) {
	r = t // if
	defer func() {
		if err := recover(); err != nil {
			r.setErr(fmt.Errorf("game stopped, event dropped"))
		}
	}()

	g.Events <- t
	return
}

type Events chan<- Event

type Event interface {
	Affect(*Game) error // called in the event loop

	// Done blocks until event is handled, can be called multiple
	// times.
	Done() error

	// setErr should set the err that Done() returnes unless already
	// done
	setErr(error)
}
