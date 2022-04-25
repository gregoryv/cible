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

func NewArea() *Area {
	return &Area{}
}

type Area struct {
	id int

	tiles []*Tile
	links []Link
}

func (me *Area) AddTile(t *Tile) {
	me.tiles = append(me.tiles, t)
}

func (me *Area) SetLinks(v []Link) error {
	for _, link := range v {
		if link.A >= me.Size() {
			return fmt.Errorf("no tile with index %d", link.A)
		}
		if link.B >= me.Size() {
			return fmt.Errorf("no tile with index %d", link.B)
		}
	}
	return nil
}

func (me *Area) Size() int { return len(me.tiles) }

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
