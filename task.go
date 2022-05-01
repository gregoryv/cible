package cible

import (
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

func Join(p Player) *EventJoin {
	return &EventJoin{
		Player: p,
	}
}
