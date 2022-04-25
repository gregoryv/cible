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
		g.EventChan() <- &EventMove{dir: East}
	})
	t.Run("handles unknown events", func(t *testing.T) {
		g := NewGame()
		go g.Run(context.Background())
		g.EventChan() <- &EventMove{dir: Direction(-1)}
	})
}
