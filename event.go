package cible

import "fmt"

type Event interface {
	Event() string
}

const (
	EventStopGame EventString = "stop game"
	EventPing     EventString = "ping"
)

type EventString string

func (me EventString) Event() string { return string(me) }

type EventMove struct {
	player Player
	dir    Direction
}

func (me *EventMove) Event() string {
	return fmt.Sprintf("%s moves %s", me.player.Name(), me.dir.String())
}
