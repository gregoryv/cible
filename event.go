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
	registerEvent(&EventLook{})

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
	// set by game
	*Character

	// set by server
	tr Transmitter
}

type EventJoin struct {
	// set by game
	Ident
}

type EventSay struct {
	Text string

	// set by server
	Ident // character who is speaking
}

type EventLeave struct {
	// set by server
	Ident
}

type Movement struct {
	Direction

	// set by server
	Ident

	// set by game
	Position
	*Tile
}

type EventLook struct {
	// set by server
	Ident // character who is looking

	Body []byte
}

func (me *Movement) String() string {
	return fmt.Sprintf("%s => %s", me.Direction, me.Position)
}

type EventStopGame struct{}
