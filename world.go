package cible

import (
	"fmt"
	"strings"
)

func Earth() World {
	return World{
		Areas: Areas{myCave()},
	}
}

func myCave() *Area {
	t1 := &Tile{
		Ident: "01",
		Short: "Cave entrance",
		Nav:   Nav{N: "02"},

		Long: `Hidden behind bushes the opening is barely visible.`,
		//
	}

	t2 := &Tile{
		Ident: "02",
		Short: "Fire room",
		Nav:   Nav{E: "03", S: "01"},

		Long: `A small streek of light comes in from a hole in the
ceiling. The entrance is a dark patch on the west wall, dryer
than the other walls.`,
		//
	}

	t3 := &Tile{
		Ident: "03",
		Short: "Small area",
		Nav:   Nav{W: "02"},
	}

	return &Area{
		Ident: "a1",
		Title: "Cave of Indy",
		Tiles: Tiles{t1, t2, t3},
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
type Short string
type Long string
type Title string
