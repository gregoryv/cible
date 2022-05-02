package cible

import (
	"context"
	"errors"
	"fmt"

	"github.com/gregoryv/logger"
)

func NewGame() *Game {
	c := NewCharactersMap()
	c.Add(&Character{
		Ident: "god",
		IsBot: true,
	})
	return &Game{
		World:      Earth(),
		MaxTasks:   10,
		Logger:     logger.Silent,
		Characters: c,
	}
}

type Game struct {
	World
	Characters

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
				if e, ok := task.Event.(fmt.Stringer); ok {
					g.Log(e.String())
				} else {
					g.Logf("%T", task.Event)
				}
			}
			// One event affects the game
			err := task.Event.AffectGame(g)
			if err != nil {
				if errors.Is(endEventLoop, err) {
					task.setErr(nil)
					break eventLoop
				}
				g.Log("event: ", err)
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

// Place returns the position as area and tile.
func (g *Game) Place(p Position) (a *Area, t *Tile, err error) {
	if a, err = g.Area(p.Area); err != nil {
		return
	}
	t, err = a.Tile(p.Tile)
	return
}

// Character returns a character in the game by id.
func (g *Game) Character(id Ident) (*Character, error) {
	return g.Characters.Character(id)
}

// ----------------------------------------

type Characters interface {
	Character(Ident) (*Character, error)
	Add(*Character)
	Remove(Ident)
	Len() int
	At(Position) []*Character
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

func (me *CharactersMap) At(p Position) []*Character {
	res := make([]*Character, 0)
	for _, c := range me.Index {
		if c.Position.Equal(p) {
			res = append(res, c)
		}
	}
	return res
}

type Character struct {
	Ident
	Name
	Position
	IsBot

	tr Transmitter //
}

func (me *Character) Transmit(m Message) error {
	return me.tr.Transmit(m)
}

func (me *Character) TransmitOthers(g *Game, m Message) error {
	nearby := g.Characters.At(me.Position)
	for _, c := range nearby {
		if c.Ident == me.Ident {
			continue
		}
		c.Transmit(m)
	}
	return nil
}

type IsBot bool

type Player struct {
	Name
}

type Bot struct{}

type Position struct {
	Area Ident
	Tile Ident
}

func (p *Position) Equal(v Position) bool {
	return p.Area == v.Area && p.Tile == v.Tile
}

type Ident string
