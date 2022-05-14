package cible

import "errors"

func NewCybromat() *Cybromat {
	return &Cybromat{
		Interactions: Interactions{
			{
				ShortAction: "i",
				Action:      "insert",
			},
		},
	}
}

type Cybromat struct {
	Interactions
}

func (me *Cybromat) InsertItem(i *Item) error {
	return ErrCannotCybernate
}

type Interactions []Interaction

type Interaction struct {
	ShortAction string
	Action      string
}

var ErrCannotCybernate = errors.New("cannot cybernate")
