package cible

import (
	"fmt"
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
		if link.From == f.T && link.Direction == dir ||
			link.To == f.T && link.Direction == dir.Opposite() {
			return area.Tiles[link.To], nil
		}
	}
	return nil, fmt.Errorf(
		"cannot move %s from %d,%d", dir.String(), f.A, f.T,
	)
}

func (me *World) Location(p At) (*Tile, error) {
	if p.A >= len(me.atlas) {
		return nil, fmt.Errorf("Location %s: area unknown", p.String())
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
