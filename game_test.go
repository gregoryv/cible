package cible

import (
	"context"
	"testing"
)

func TestGame(t *testing.T) {
	t.Run("can be cancelled", func(t *testing.T) {
		g := NewGame()
		ctx, cancel := context.WithCancel(context.Background())
		go g.Run(ctx)
		cancel()
	})
	t.Run("can be stopped", func(t *testing.T) {
		g := NewGame()
		go g.Run(context.Background())
		g.EventChan() <- EventStopGame
	})
	t.Run("handles events", func(t *testing.T) {
		g := NewGame()
		go g.Run(context.Background())
		g.EventChan() <- EventPing
		g.EventChan() <- EventPing
	})

}

type EventString string

func (me EventString) Event() string { return string(me) }
