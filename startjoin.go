package cible

import (
	"context"
	"fmt"
)

func (me *Game) Run(ctx context.Context) error {
gameLoop:
	for {
		select {
		case <-ctx.Done(): // ie. interrupted from the outside
			break gameLoop

		case e := <-me.events: // blocks
			switch e := e.(type) {
			case *EventStopGame:
				me.Log(e.Event())
				e.failed <- nil //
				break gameLoop

			default:
				if err := me.handleEvent(e); err != nil {
					me.Logf("%s: %v", e.Event(), err)
				} else {
					me.Log(e.Event())
				}
			}
		}
	}

	return nil
}

func (me *Game) handleEvent(e Event) error {
	switch e := e.(type) {
	case *EventJoin:
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

	case *EventLeave:
		c, err := me.Character(e.Ident)
		if err != nil {
			e.failed <- err
			return err
		}
		e.Name = c.Player.Name
		e.failed <- nil
		return nil

	case *Movement:
		return me.onMovement(e)

	}
	return nil
}

func Trigger[T Event](g *Game, t T) T {
	g.Events <- t
	return t
}

// ----------------------------------------
func (me *Game) Stop() { me.Events <- StopGame() }

func StopGame() *EventStopGame {
	return &EventStopGame{
		failed: make(chan error, 1),
	}
}

type EventStopGame struct {
	failed chan error
}

func (me *EventStopGame) Event() string {
	return "stop game"
}

func (me *EventStopGame) Done() error {
	defer me.Close()
	return <-me.failed
}

func (me *EventStopGame) Close() {
	close(me.failed)
}

// ----------------------------------------

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

func (me *EventLeave) Close() { close(me.failed) }

func (me *EventLeave) Event() string {
	return fmt.Sprintf("%s left", me.Name)
}

// ----------------------------------------

type Events chan<- Event

type Event interface {
	Event() string
}
