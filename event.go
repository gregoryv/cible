package cible

import (
	"fmt"
	"sync"
)

func TriggerWait[T Event](g *Game, e T) (r T) {
	task, r := Trigger(g, e)
	task.Done()
	return
}

// Trigger an event which affects the game. Callers should call Done
// to wait for the event.
func Trigger[T Event](g *Game, e T) (task *Task, r T) {
	fmt.Printf("triggering: %#v", e)
	r = e // during panic t, would be nil
	defer func() {
		if err := recover(); err != nil {
			task.setErr(fmt.Errorf("game stopped, event dropped"))
		}
	}()

	task = NewTask(e)
	g.ch <- task
	return
}

type Event interface {
	Affect(*Game) error // called in the event loop
}

func NewTask(e Event) *Task {
	return &Task{
		Event:  e,
		failed: make(chan error, 1),
	}
}

type Task struct {
	Event

	err error

	once   sync.Once
	failed chan error
}

// Done blocks until event is handled, can be called multiple
// times.
func (me *Task) Done() error {
	me.once.Do(func() {
		select {
		case me.err = <-me.failed:
		}
		close(me.failed)
	})
	return me.err
}

func (me *Task) setErr(v error) {
	me.failed <- v
}
