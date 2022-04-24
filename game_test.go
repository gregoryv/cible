package cible

import (
	"context"
	"testing"
)

func TestGame(t *testing.T) {
	g := NewGame()
	ctx, cancel := context.WithCancel(context.Background())
	go g.Run(ctx)
	g.EventChan() <- EventString("hello")
	cancel()
}

type EventString string

func (me EventString) Event() string { return string(me) }
