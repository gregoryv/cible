package cible

import (
	"fmt"
	"sync"
)

func MoveCharacter(id Ident, d Direction) *Movement {
	return &Movement{
		Ident:       id,
		Direction:   d,
		newPosition: make(chan Position, 1),
		failed:      make(chan error, 1),
	}
}

type Movement struct {
	Ident
	Direction

	Position // set when done
	err      error

	sync.Once
	newPosition chan Position
	failed      chan error
}

func (e *Movement) Affect(g *Game) (err error) {
	g.Logf("%s move %s", e.Ident, e.Direction)
	defer func() {
		if err != nil {
			e.failed <- err
		}
	}()
	c, err := g.Character(e.Ident)
	// always send a position as someone might be waiting for a
	// response
	defer func() {
		if c == nil {
			return
		}
		e.newPosition <- c.Position
	}()
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
	return nil
}

func (me *Movement) Done() (err error) {
	me.Once.Do(func() {
		select {
		case me.Position = <-me.newPosition:
		case me.err = <-me.failed:
		}
		close(me.newPosition)
		close(me.failed)
	})
	return me.err
}

func (e *Movement) setErr(v error) {
	e.err = v
	close(e.failed)
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
