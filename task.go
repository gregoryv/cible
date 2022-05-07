package cible

import (
	"fmt"
	"sync"
)

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

func (me *Task) String() string {
	if e, ok := me.Event.(fmt.Stringer); ok {
		return fmt.Sprintf("%T %s", me.Event, e.String())
	}
	return fmt.Sprintf("%T", me.Event)
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

// gomerge src: joinleave.go

func Join(p Player) *PlayerJoin {
	return &PlayerJoin{
		Player: p,
	}
}
