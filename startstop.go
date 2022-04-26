package cible

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

func (me *EventStopGame) Done() (err error) {
	defer me.Close()
	select {
	case err = <-me.failed:
	}
	return
}

func (me *EventStopGame) Close() {
	defer ignorePanic()
	close(me.failed)
}
