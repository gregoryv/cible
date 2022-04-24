package cible

import (
	"context"
	"time"
)

func NewGame() *Game {
	return &Game{}
}

type Game struct{}

func (me *Game) Run(ctx context.Context) error {

gameLoop:
	for {
		select {
		case <-ctx.Done(): // ie. interrupted from the outside
			break gameLoop

		default:
			// do stuff

			<-time.After(time.Second) // todo remove later
		}
	}

	return nil
}
