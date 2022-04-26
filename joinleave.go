package cible

import "fmt"

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

	joined chan Ident
	failed chan error
}

func (me *EventJoin) Done() (id Ident, err error) {
	defer me.Close()
	select {
	case id = <-me.joined:
	case err = <-me.failed:
	}
	return
}

func (me *EventJoin) Close() {
	close(me.joined)
	close(me.failed)
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
	failed chan error
}

func (me *EventLeave) Done() (err error) {
	defer me.Close()
	return <-me.failed
	return
}

func (me *EventLeave) Close() {
	defer ignorePanic()
	close(me.failed)
}

func (me *EventLeave) Event() string {
	return fmt.Sprintf("%s left", me.Name)
}
