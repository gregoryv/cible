package cible

import (
	"encoding/gob"
	"fmt"
	"log"
)

// Events follow a command pattern so we can send events accross the
// wire using some encoding.
type Event interface{}

type GameEvent interface {
	AffectGame(*Game) error
}

type UIEvent interface {
	AffectUI(*UI)
}

// ----------------------------------------

func init() { registerEvent(&EventJoin{}) }

type EventJoin struct {
	Player
	*Character

	tr Transmitter
}

func (e *EventJoin) AffectGame(g *Game) error {
	c := &Character{
		Name: e.Player.Name,
		Position: Position{
			Area: "a1", Tile: "01",
		},
		tr: e.tr,
	}
	g.Characters.Add(c)
	g.Logf("%s joined game as %s", c.Name, c.Ident)
	e.Character = c

	// notify others of the new character
	go c.TransmitOthers(g,
		NewMessage(&CharacterJoined{Ident: c.Ident}),
	)
	return c.Transmit(NewMessage(e)) // back to player
}

func init() { registerEvent(&CharacterJoined{}) }

type CharacterJoined struct {
	Ident
}

func (e *CharacterJoined) AffectUI(ui *UI) {
	ui.OtherPlayer(e.Ident, "joined")
}

// ----------------------------------------

func init() { registerEvent(&EventSay{}) }

type EventSay struct {
	Ident // character who is speaking
	Text  string
}

func (e *EventSay) AffectGame(g *Game) error {
	c, err := g.Characters.Character(e.Ident)
	if err != nil {
		return err
	}
	go c.TransmitOthers(g, NewMessage(e))
	return nil
}

func (e *EventSay) AffectUI(ui *UI) {
	ui.OtherPlayerSays(e.Ident, e.Text)
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

func (e *EventLeave) AffectGame(g *Game) error {
	c, err := g.Characters.Character(e.Ident)
	if err != nil {
		return err
	}
	g.Characters.Remove(c.Ident)
	g.Logf("%s left, %v remaining", c.Name, g.Characters.Len())
	go c.TransmitOthers(g, NewMessage(e))
	return nil
}

func (e *EventLeave) AffectUI(ui *UI) {
	ui.OtherPlayer(e.Ident, "left game")
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

func (e *Movement) AffectGame(g *Game) (err error) {
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

// Do Not register this event as it would allow a client to stop the
// server.

type EventStopGame struct{}

func (e *EventStopGame) AffectGame(g *Game) error {
	// special event that ends the loop, thus we do things here as
	// no other events should be affecting the game
	g.Log("shutting down...")

	return endEventLoop
}

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
