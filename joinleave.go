package cible

import (
	"fmt"
	"sync"
)

func (me *Game) onJoin(e *EventJoin) error {
	p := Position{
		Area: "a1", Tile: "01",
	}
	me.Characters = append(me.Characters, &Character{
		Ident:    Ident(e.Player.Name),
		Player:   e.Player,
		Position: p,
	})
	e.joined <- Ident(e.Player.Name)
	return nil
}

func Join(p Player) *EventJoin {
	return &EventJoin{
		Player: p,
		joined: make(chan Ident, 1), // buffer so event loop doesn't block
		failed: make(chan error, 1),
	}
}

type EventJoin struct {
	Player

	Ident // set when done
	err   error

	sync.Once
	joined chan Ident
	failed chan error
}

func (me *EventJoin) Done() (err error) {
	me.Once.Do(func() {
		select {
		case me.Ident = <-me.joined:
		case me.err = <-me.failed:
		}
		close(me.joined)
		close(me.failed)
	})
	return me.err
}

func (me *EventJoin) Event() string {
	return fmt.Sprintf("%s join", me.Player.Name)
}

// ----------------------------------------

func (me *Game) onLeave(e *EventLeave) error {
	c, err := me.Character(e.Ident)
	if err != nil {
		e.failed <- err
		return err
	}
	e.Name = c.Player.Name
	e.failed <- nil
	return nil
}

func Leave(cid Ident) *EventLeave {
	return &EventLeave{
		Ident:  cid,
		failed: make(chan error, 1),
	}
}

type EventLeave struct {
	Ident
	Name

	err error // set when done

	sync.Once
	failed chan error
}

func (me *EventLeave) Done() (err error) {
	me.Once.Do(func() {
		me.err = <-me.failed
		close(me.failed)
	})
	return me.err
}

func (me *EventLeave) Event() string {
	return fmt.Sprintf("%s left", me.Name)
}
