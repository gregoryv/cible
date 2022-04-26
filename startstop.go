package cible

import "sync"

func (me *Game) Stop() { me.Events <- StopGame() }

func StopGame() *EventStopGame {
	return &EventStopGame{
		failed: make(chan error, 1),
	}
}

type EventStopGame struct {
	err error // set when done

	sync.Once
	failed chan error
}

func (me *EventStopGame) Event() string {
	return "stop game"
}

func (me *EventStopGame) Done() error {
	me.Once.Do(func() {
		me.err = <-me.failed
		close(me.failed)
	})
	return me.err
}
