package cible

import (
	"context"
	"errors"
	"fmt"

	"github.com/gregoryv/logger"
)

func NewGame() *Game {
	return &Game{
		World:      NewWorld(),
		Characters: NewCharactersMap(),
		Items: Items{
			{
				Name:     "ball",
				Count:    1,
				Location: Location{Area: "a1", Tile: "t9"},
			},
		},
		MaxTasks: 10,
		Logger:   logger.Silent,
	}
}

type Game struct {
	World
	Characters
	Items

	MaxTasks     int
	LogAllEvents bool

	ch chan *Task
	logger.Logger
}

func (g *Game) Run(ctx context.Context) error {
	g.Log("start game")
	g.ch = make(chan *Task, g.MaxTasks)

eventLoop:
	for {
		select {
		case <-ctx.Done(): // ie. interrupted from the outside
			break eventLoop

		case task := <-g.ch: // blocks
			if g.LogAllEvents {
				g.Log(task.String())
			}
			// One event affects the game
			err := g.AffectGame(task.Event)
			if err != nil {
				if errors.Is(endEventLoop, err) {
					task.setErr(nil)
					break eventLoop
				}
				g.Logf("%T %v", task.Event, err)
			}

			// Make sure any event can be cleaned up. Triggering
			// side will most likely also wait for event to be
			// done, but this is here to give them the option to
			// ignore it. This does impact performance quite a bit
			// though.
			task.setErr(err)
		}
	}
	close(g.ch)
	g.Log("game stopped")
	return nil
}

func (g *Game) AffectGame(e interface{}) error {
	switch e := e.(type) {

	case *EventSay:
		c, err := g.Characters.Character(e.Ident)
		if err != nil {
			return err
		}
		e.Name = c.Name
		go c.TransmitOthers(g, NewMessage(e))

	case *EventJoin:

	case *EventJoinGame:
		c := &Character{
			Name: e.Player.Name,
			Location: Location{
				Area: "a1", Tile: "t1",
			},
			Inventory: *NewInventory(),
			tr:        e.tr,
		}
		g.Characters.Add(c)
		g.Logf("%s joined game as %s", c.Name, c.Ident)
		e.Character = c
		a, _, _ := g.Place(c.Location)
		e.Title = a.Title

		// notify others of the new character
		go c.TransmitOthers(g,
			NewMessage(&EventJoin{
				Ident: c.Ident,
				Name:  c.Name,
			}),
		)
		return c.Transmit(NewMessage(e)) // back to player

	case *EventLeave:
		c, err := g.Characters.Character(e.Ident)
		if err != nil {
			return err
		}
		e.Name = c.Name
		g.Characters.Remove(c.Ident)
		g.Logf("%s left, %v remaining", c.Name, g.Characters.Len())
		go c.TransmitOthers(g, NewMessage(e))

	case *EventMove:
		g.Logf("%s move %s", e.Ident, e.Direction)
		c, err := g.Character(e.Ident)
		if err != nil {
			return err
		}

		_, t, err := g.Place(c.Location)
		if err != nil {
			return err
		}
		next, err := link(t, e.Direction)
		if err != nil {
			return err
		}
		// must do this Before setting next position
		c.TransmitOthers(g, NewMessage(&EventGoAway{Name: c.Name}))

		if next != "" {
			c.Location.Tile = next
		}
		e.Location = c.Location
		a, t, _ := g.Place(c.Location)
		e.Tile = t
		e.Title = a.Title
		e.Body = []byte(t.Short + "...")
		go c.Transmit(NewMessage(e))
		go c.TransmitOthers(g, NewMessage(&EventApproach{Name: c.Name}))

	case *EventLook:
		c, err := g.Character(e.Ident)
		if err != nil {
			return err
		}
		_, t, err := g.Place(c.Location)
		if err != nil {
			return err
		}

		e.Tile = *t
		e.Loose = g.Items.At(c.Location)
		go c.Transmit(NewMessage(e))

	case *EventExamine:
		c, err := g.Character(e.Ident)
		if err != nil {
			return err
		}
		_, t, err := g.Place(c.Location)
		if err != nil {
			return err
		}
		if t.Cybromat != nil && e.Item.Name == "cybromat" {
			e.Interactions = t.Cybromat.Interactions
		} else {
			e.Note = fmt.Sprintf("cannot examine %s", e.Item.Name)
		}
		go c.Transmit(NewMessage(e))

	case *EventPickup:
		c, err := g.Character(e.Ident)
		if err != nil {
			return err
		}
		item, err := g.Items.At(c.Location).FindByName(e.Item.Name)
		if err != nil {
			c.Transmit(NewMessage(e))
			return nil
		}
		item.Location.Area = ""
		item.Location.Tile = ""
		e.Item.Count = 1
		c.Inventory.AddItem(e.Item)
		c.Transmit(NewMessage(&EventInventoryUpdate{&c.Inventory}))

	case *EventStopGame:
		// special event that ends the loop, thus we do things here as
		// no other events should be affecting the game
		g.Log("shutting down...")

		return endEventLoop

	case *EventDisconnect:
		// todo

	case interface{ AffectGame(*Game) error }:
		return e.AffectGame(g)

	default:
		return fmt.Errorf("unknown event %T", e)
	}
	return nil
}

// Do enques the task and waits for it to complete
func (g *Game) Do(e Event) error {
	t := NewTask(e)
	g.Enqueue(t)
	t.Done()
	return t.err
}

func (g *Game) Enqueue(t *Task) {
	defer func() {
		// handle closed channel, ie. game stopped
		if err := recover(); err != nil {
			t.setErr(fmt.Errorf("game stopped, event dropped"))
		}
	}()
	g.ch <- t
}

// Place returns the Location as area and tile.
func (g *Game) Place(loc Location) (a *Area, t *Tile, err error) {
	if a, err = g.Area(loc.Area); err != nil {
		return
	}
	t, err = a.Tile(loc.Tile)
	return
}

// Character returns a character in the game by id.
func (g *Game) Character(id Ident) (*Character, error) {
	return g.Characters.Character(id)
}

var endEventLoop = fmt.Errorf("end event loop")

func link(t *Tile, d Direction) (Ident, error) {
	if d < 0 || int(d) > len(t.Nav) {
		return "", fmt.Errorf("bad direction")
	}
	return t.Nav[int(d)], nil
}

// ----------------------------------------

type Characters interface {
	Character(Ident) (*Character, error)
	Add(*Character)
	Remove(Ident)
	Len() int
	At(Location) []*Character
}

func NewCharactersMap() *CharactersMap {
	return &CharactersMap{
		Index: make(map[Ident]*Character),
	}
}

type CharactersMap struct {
	Index   map[Ident]*Character
	idCount int
}

func (me *CharactersMap) Character(id Ident) (*Character, error) {
	c, found := me.Index[id]
	if !found {
		return nil, fmt.Errorf("character %q not found", id)
	}
	return c, nil
}

func (me *CharactersMap) Add(c *Character) {
	me.idCount++
	c.Ident = Ident(fmt.Sprintf("char%02v", me.idCount))
	me.Index[c.Ident] = c
}

func (me *CharactersMap) Remove(id Ident) {
	delete(me.Index, id)
}

func (me *CharactersMap) Len() int {
	return len(me.Index)
}

func (me *CharactersMap) At(loc Location) []*Character {
	res := make([]*Character, 0)
	for _, c := range me.Index {
		if c.Location.Equal(loc) {
			res = append(res, c)
		}
	}
	return res
}
