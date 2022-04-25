package cible

import (
	"fmt"
	"log"
)

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

func (me *World) Move(dir Direction, f From) (*Tile, error) {
	if _, err := me.Location(f); err != nil {
		return nil, err
	}
	area := me.atlas[f.A]
	for _, link := range area.Links {
		log.Println("link", link)
		if link.From == f.T && link.Direction == dir {
			return area.Tiles[link.To], nil
		}
	}
	return nil, fmt.Errorf(
		"cannot move %s from %d,%d", dir.String(), f.A, f.T,
	)
}

func (me *World) Location(p At) (*Tile, error) {
	if p.A >= len(me.atlas) {
		return nil, fmt.Errorf("Location %s: unknown", p.String())
	}
	return me.atlas[p.A].Location(p)
}

type From = At

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
		if link.From >= me.Size() {
			return fmt.Errorf("no tile with index %d", link.From)
		}
		if link.To >= me.Size() {
			return fmt.Errorf("no tile with index %d", link.To)
		}
		mark[link.From] = true
		mark[link.To] = true
	}
	for i, v := range mark {
		if !v {
			return fmt.Errorf("missing link for %v", i)
		}
	}
	me.Links = v
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
	From, To int
	Direction
}

func (me *Link) String() string {
	return fmt.Sprintf("%d %v %d", me.From, me.Direction, me.To)
}
