package cible

import (
	"encoding/gob"
	"fmt"
	"log"
)

func NewEvent(name string) (Event, bool) {
	if fn, found := eventConstructors[name]; !found {
		log.Println(name, "NOT REGISTERED")
		return nil, false
	} else {
		return fn(), true
	}
}

type Event interface{}

var eventConstructors = make(map[string]func() Event)

func init() {
	registerEvent(&PlayerJoin{})
	registerEvent(&EventJoin{})
	registerEvent(&EventSay{})
	registerEvent(&EventLeave{})
	registerEvent(&Movement{})

	// Do Not register EventStopGame as it would allow a client to
	// stop the server.
}

// register pointer to events
func registerEvent[T Event](t *T) {
	gob.Register(*t)
	eventConstructors[fmt.Sprintf("%T", *t)] = func() Event {
		var x T
		return &x
	}
}

// ----------------------------------------

type PlayerJoin struct {
	Player
	*Character // populated by game

	tr Transmitter // populated by server
}

type EventJoin struct {
	Ident
}

type EventSay struct {
	Ident // character who is speaking
	Text  string
}

func Leave(cid Ident) *EventLeave {
	return &EventLeave{
		Ident: cid,
	}
}

type EventLeave struct {
	Ident
}

func MoveCharacter(id Ident, d Direction) *Movement {
	return &Movement{
		Ident:     id,
		Direction: d,
	}
}

type Movement struct {
	Ident
	Direction

	Position // set by game
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

type EventStopGame struct{}

// ----------------------------------------

var endEventLoop = fmt.Errorf("end event loop")
