package cible

import (
	"fmt"
	"sync"
)

func TriggerWait[T Event](g *Game, e T) (r T) {
	task, r := Trigger(g, e)
	task.Done()
	return
}

// Trigger an event which affects the game. Callers should call Done
// to wait for the event.
func Trigger[T Event](g *Game, e T) (task *Task, r T) {
	r = e // during panic t, would be nil
	defer func() {
		if err := recover(); err != nil {
			task.setErr(fmt.Errorf("game stopped, event dropped"))
		}
	}()

	task = NewTask(e)
	g.ch <- task
	return
}

type Event interface {
	Affect(*Game) error // called in the event loop
}

func NewTask(e Event) *Task {
	return &Task{
		Event:  e,
		failed: make(chan error, 1),
	}
}

type Task struct {
	Event

	err error

	once   sync.Once
	failed chan error
}

// Done blocks until event is handled, can be called multiple
// times.
func (me *Task) Done() error {
	me.once.Do(func() {
		select {
		case me.err = <-me.failed:
		}
		close(me.failed)
	})
	return me.err
}

func (me *Task) setErr(v error) {
	me.failed <- v
}

// gomerge src: joinleave.go

func Join(p Player) *EventJoin {
	return &EventJoin{
		Player: p,
	}
}

type EventJoin struct {
	Player
	Ident // set when done
}

func (e *EventJoin) Affect(g *Game) error {
	g.Logf("%s join", e.Player.Name)
	p := Position{
		Area: "a1", Tile: "01",
	}
	c := &Character{
		Ident:    Ident(e.Player.Name),
		Name:     e.Player.Name,
		Position: p,
	}
	g.Characters = append(g.Characters, c)
	e.Ident = c.Ident
	return nil
}

// ----------------------------------------

func Leave(cid Ident) *EventLeave {
	return &EventLeave{
		Ident: cid,
	}
}

type EventLeave struct {
	Ident
}

func (e *EventLeave) Affect(g *Game) error {
	c, err := g.Character(e.Ident)
	if err != nil {
		return err
	}
	g.Logf("%s left", c.Name)
	return nil
}

// gomerge src: movement.go

func MoveCharacter(id Ident, d Direction) *Movement {
	return &Movement{
		Ident:     id,
		Direction: d,
	}
}

type Movement struct {
	Ident
	Direction

	Position // set when done
}

func (e *Movement) Affect(g *Game) (err error) {
	g.Logf("%s move %s", e.Ident, e.Direction)
	c, err := g.Character(e.Ident)
	if err != nil {
		return err
	}

	_, t, err := g.Place(c.Position)
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
	e.Position = c.Position
	return nil
}

func (me *Tile) Link(d Direction) (Ident, error) {
	if d < 0 || int(d) > len(me.Nav) {
		return "", fmt.Errorf("bad direction")
	}
	return me.Nav[int(d)], nil
}

// ----------------------------------------

type Direction int

const (
	N Direction = iota
	E
	S
	W
)

// gomerge src: startstop.go

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
