package cible

// Trigger an event which affects the game. Callers should call Done
// to wait for the event.
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
