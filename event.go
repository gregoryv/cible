package cible

import (
	"encoding/gob"
	"fmt"
	"log"
)

func init() {
	registerEvent(&EventJoinGame{})
	registerEvent(&EventJoin{})
	registerEvent(&EventSay{})
	registerEvent(&EventLeave{})
	registerEvent(&EventMove{})
	registerEvent(&EventLook{})
	registerEvent(&EventDisconnect{})
	registerEvent(&EventApproach{})
	registerEvent(&EventGoAway{})
	registerEvent(&EventPickup{})
	registerEvent(&EventExamine{})
	registerEvent(&EventInventoryUpdate{})

	// Do Not register EventStopGame as it would allow a client to
	// stop the server.
}

// ----------------------------------------

type EventInventoryUpdate struct {
	*Inventory
}

type EventExamine struct {
	Ident
	Item

	Interactions
	Note string
}

type EventPickup struct {
	Ident
	Item

	ItemFound
}

type EventGoAway struct {
	Name
}

// when character enters a tile
type EventApproach struct {
	Name
}

type EventJoinGame struct {
	Player
	// set by game
	*Character

	// set by server
	tr Transmitter

	Title // of the area
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

	// set by game
	Name
}

type EventLeave struct {
	// set by server
	Ident

	// set by game
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
	Location
	Title // of the area
	*Tile
	Body []byte
}

func (me *EventMove) String() string {
	return fmt.Sprintf("%s => %s", me.Direction, me.Location)
}

type EventLook struct {
	// set by server
	Ident // character who is looking

	Tile

	Loose Items
}

type EventStopGame struct{}

// ----------------------------------------

// NewEvent returns a new instance of the named event. Returns false
// if the event has not been registered.
func NewEvent(name string) (Event, bool) {
	if fn, found := eventConstructors[name]; !found {
		log.Println(name, "NOT REGISTERED")
		return nil, false
	} else {
		return fn(), true
	}
}

type Event interface{}

// register pointer to events
func registerEvent[T Event](t *T) {
	gob.Register(*t)
	eventConstructors[fmt.Sprintf("%T", *t)] = func() Event {
		var x T
		return &x
	}
}

var eventConstructors = make(map[string]func() Event)
