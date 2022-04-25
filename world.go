package cible

import "fmt"

func NewWorld() *World {
	return &World{
		atlas: make([]*Area, 0),
	}
}

type World struct {
	atlas []*Area
}

func (me *World) AddArea(v *Area) {
	me.atlas = append(me.atlas, v)
}

func (me *World) Location(p At) (*Tile, error) {
	if p.A >= len(me.atlas) {
		return nil, fmt.Errorf("Location %s: unknown", p.String())
	}
	return me.atlas[p.A].Location(p)
}

type At struct {
	A int // Area id
	T int // Tile id
}

func (me *At) String() string {
	return fmt.Sprintf("at %d,%d", me.A, me.T)
}

func NewArea() *Area {
	return &Area{}
}

type Area struct {
	Tiles []*Tile
	Links []Link
}

func (me *Area) Location(p At) (*Tile, error) {
	if p.T >= me.Size() {
		return nil, fmt.Errorf("Location %v: unknown", p.String())
	}
	return me.Tiles[p.T], nil
}

func (me *Area) AddTile(t *Tile) {
	me.Tiles = append(me.Tiles, t)
}

func (me *Area) SetLinks(v []Link) error {
	mark := make([]bool, len(me.Tiles))

	for _, link := range v {
		if link.A >= me.Size() {
			return fmt.Errorf("no tile with index %d", link.A)
		}
		if link.B >= me.Size() {
			return fmt.Errorf("no tile with index %d", link.B)
		}
		mark[link.A] = true
		mark[link.B] = true
	}
	for i, v := range mark {
		if !v {
			return fmt.Errorf("missing link for %v", i)
		}
	}
	return nil
}

func (me *Area) Size() int { return len(me.Tiles) }

func NewTile() *Tile {
	return &Tile{}
}

type Tile struct {
	Short []byte
	Long  []byte
}

type Link struct {
	A, B int
	Direction
}
