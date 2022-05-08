package cible

import (
	"encoding/gob"
	"fmt"
	"log"
)

func init() {
	registerEvent(&PlayerJoin{})
	registerEvent(&EventJoin{})
	registerEvent(&EventSay{})
	registerEvent(&EventLeave{})
	registerEvent(&EventMove{})
	registerEvent(&EventLook{})
	registerEvent(&EventDisconnect{})
	registerEvent(&EventRenameCharacter{})

	// Do Not register EventStopGame as it would allow a client to
	// stop the server.
}

// ----------------------------------------

type EventRenameCharacter struct {
	Ident // current
}

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
	Name
}

type EventSay struct {
	Text string

	// set by server, character who is speaking
	Ident
	Name
}

type EventLeave struct {
	// set by server
	Ident
	Name
}

type EventDisconnect struct {
	// set by server
	Ident
}

// Your character EventMove in the game
type EventMove struct {
	Direction

	// set by server
	Ident

	// set by game
	Position
	*Tile
	Body []byte
}

func (me *EventMove) String() string {
	return fmt.Sprintf("%s => %s", me.Direction, me.Position)
}

type EventLook struct {
	// set by server
	Ident // character who is looking

	Tile
}

type EventStopGame struct{}

// ----------------------------------------

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

// register pointer to events
func registerEvent[T Event](t *T) {
	gob.Register(*t)
	eventConstructors[fmt.Sprintf("%T", *t)] = func() Event {
		var x T
		return &x
	}
}
