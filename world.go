package cible

import (
	"fmt"
	"strings"
)

func NewWorld() World {
	return World{
		Areas: Areas{Spaceport()},
	}
}

type World struct {
	Areas
}

type Areas []*Area

func (me Areas) Area(id Ident) (*Area, error) {
	for _, a := range me {
		if a.Ident == id {
			return a, nil
		}
	}
	return nil, fmt.Errorf("area %q not found", id)
}

type Area struct {
	Ident
	Title
	Tiles
}

func (a *Area) Tile(id Ident) (*Tile, error) {
	for _, t := range a.Tiles {
		if t.Ident == id {
			return t, nil
		}
	}
	return nil, fmt.Errorf("tile %q not found", id)
}

func (a *Area) AddTile(tiles ...*Tile) {
	for _, t := range tiles {
		a.Tiles = append(a.Tiles, t)
		t.Ident.SetIdent(fmt.Sprintf("t%d", len(a.Tiles)))
	}
}

type Tiles []*Tile

// Nav references tiles based on direction Direction -> Tile.Ident
type Nav [8]Ident

func (n Nav) String() string {
	var res []string
	for d, id := range n {
		if id != "" {
			res = append(res, Direction(d).String()+":"+string(id))
		}
	}
	return strings.Join(res, " ")
}

//go:generate stringer -output direction_string.go -type Direction
type Direction int

const (
	N Direction = iota
	NE
	E
	SE
	S
	SW
	W
	NW
)

var opposite = [8]Direction{
	N:  S,
	NE: SW,
	E:  W,
	SE: NW,
	S:  N,
	SW: NE,
	W:  E,
	NW: SE,
}
