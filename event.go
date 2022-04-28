package cible

import (
	"encoding/gob"
	"fmt"
)

// Events follow a command pattern so we can send events accross the
// wire using some encoding.
type Event interface {
	Affect(*Game) error // called in the event loop
}

type EventJoin struct {
	Player
	Ident // set when done
}

func (e *EventJoin) Affect(g *Game) error {
	p := Position{
		Area: "a1", Tile: "01",
	}
	c := &Character{
		Name:     e.Player.Name,
		Position: p,
	}
	g.Characters.Add(c)
	e.Ident = c.Ident
	g.Logf("%s joined game as %s", e.Player.Name, c.Ident)
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

// ----------------------------------------

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
	*Tile
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
	next, err := link(t, e.Direction)
	if err != nil {
		return err
	}
	if next != "" {
		c.Position.Tile = next
	}
	e.Position = c.Position
	e.Tile = t
	return nil
}

func (me *Movement) String() string {
	return fmt.Sprintf("%s => %s", me.Direction, me.Position)
}

func link(t *Tile, d Direction) (Ident, error) {
	if d < 0 || int(d) > len(t.Nav) {
		return "", fmt.Errorf("bad direction")
	}
	return t.Nav[int(d)], nil
}

// ----------------------------------------

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

// ----------------------------------------

var endEventLoop = fmt.Errorf("end event loop")

// value must be interface{}, but also implement Event
func newNamedEvent(name string) (interface{}, bool) {
	if fn, found := eventConstructors[name]; !found {
		return nil, false
	} else {
		return fn(), true
	}
}

func init() {
	// register events that remote clients could send
	registerEvent(&EventJoin{})
	registerEvent(&Movement{})
}

// register pointer to events
func registerEvent[T any](t *T) {
	gob.Register(*t)
	eventConstructors[fmt.Sprintf("%T", *t)] = func() interface{} {
		var x T
		return &x
	}
}

var eventConstructors = make(map[string]func() interface{})
