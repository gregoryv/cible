package cible

import (
	"fmt"
	"log"
	"strings"
)

func Earth() World {
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

type Tile struct {
	Ident
	Short
	Long
	Nav
}

func (t *Tile) String() string {
	return fmt.Sprintf("%s %s", t.Ident, t.Short)
}

func (me *Tile) Link(t *Tile, d Direction) {
	if me.Nav[d] != "" {
		panic(
			fmt.Sprintf(
				"cannot link %s, %s already linked to %v",
				me.String(), d.String(), me.Nav[d],
			),
		)
	}
	log.Printf("me.Nav[%s]=%s", d, me.Nav[d])
	me.Nav[d] = t.Ident
}

type Nav [4]Ident

func (n Nav) String() string {
	var res []string
	for d, id := range n {
		if id != "" {
			res = append(res, Direction(d).String()+":"+string(id))
		}
	}
	return strings.Join(res, " ")
}

type Direction int

const (
	N Direction = iota
	E
	S
	W
)

type Name string

func (me *Name) SetName(v string) {
	*me = Name(v)
}

type Short string
type Long string
type Title string
