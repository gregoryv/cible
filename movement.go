package cible

import "fmt"

func (me *Game) onMovement(e *Movement) (err error) {
	defer func() {
		if err != nil {
			e.failed <- err
		}
	}()
	c, err := me.Character(e.Ident)
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

	_, t, err := me.Place(c.Position)
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

func (me *Tile) Link(d Direction) (Ident, error) {
	if d < 0 || int(d) > len(me.Nav) {
		return "", fmt.Errorf("bad direction")
	}
	return me.Nav[int(d)], nil
}

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

	newPosition chan Position
	failed      chan error
}

func (me *Movement) Done() (pos Position, err error) {
	defer me.Close()
	select {
	case pos = <-me.newPosition:
	case err = <-me.failed:
	}
	return
}

func (me *Movement) Close() {
	defer ignorePanic()
	close(me.newPosition)
	close(me.failed)
}

func (me *Movement) Event() string {
	return fmt.Sprintf("%s move %s", me.Ident, me.Direction)
}

func ignorePanic() { _ = recover() }

// ----------------------------------------

type Direction int

const (
	N Direction = iota
	E
	S
	W
)
