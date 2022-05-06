package cible

import (
	"encoding/gob"
	"fmt"
	"log"
)

// Events follow a command pattern so we can send events accross the
// wire using some encoding.
type Event interface{}

// ----------------------------------------

func init() { registerEvent(&EventJoin{}) }

type EventJoin struct {
	Player
	*Character

	tr Transmitter // populated by server
}

func init() { registerEvent(&CharacterJoin{}) }

type CharacterJoin struct {
	Ident
}

// ----------------------------------------

func init() { registerEvent(&EventSay{}) }

type EventSay struct {
	Ident // character who is speaking
	Text  string
}

// ----------------------------------------

func Leave(cid Ident) *EventLeave {
	return &EventLeave{
		Ident: cid,
	}
}

func init() { registerEvent(&EventLeave{}) }

type EventLeave struct {
	Ident
}

// ----------------------------------------

func MoveCharacter(id Ident, d Direction) *Movement {
	return &Movement{
		Ident:     id,
		Direction: d,
	}
}

func init() { registerEvent(&Movement{}) }

type Movement struct {
	Ident
	Direction

	Position // set when done
	*Tile
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

// Do Not register this event as it would allow a client to stop the
// server.

type EventStopGame struct{}

// ----------------------------------------

var endEventLoop = fmt.Errorf("end event loop")

// value must be interface{}, but also implement Event
func NewNamedEvent(name string) (interface{}, bool) {
	if fn, found := eventConstructors[name]; !found {
		log.Println(name, "NOT REGISTERED")
		return nil, false
	} else {
		return fn(), true
	}
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
