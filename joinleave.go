package cible

import (
	"sync"
)

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

func (e *EventJoin) Affect(g *Game) error {
	g.Logf("%s join", e.Player.Name)
	p := Position{
		Area: "a1", Tile: "01",
	}
	g.Characters = append(g.Characters, &Character{
		Ident:    Ident(e.Player.Name),
		Name:     e.Player.Name,
		Position: p,
	})
	e.joined <- Ident(e.Player.Name)
	return nil
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

// ----------------------------------------

func Leave(cid Ident) *EventLeave {
	return &EventLeave{
		Ident:  cid,
		failed: make(chan error, 1),
	}
}

type EventLeave struct {
	Ident

	err error // set when done

	sync.Once
	failed chan error
}

func (e *EventLeave) Affect(g *Game) error {
	c, err := g.Character(e.Ident)
	if err != nil {
		e.failed <- err
		return err
	}
	e.failed <- nil
	g.Logf("%s left", c.Name)
	return nil
}

func (me *EventLeave) Done() (err error) {
	me.Once.Do(func() {
		me.err = <-me.failed
		close(me.failed)
	})
	return me.err
}
