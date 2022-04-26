package cible

import (
	"fmt"
)

func NewArea() *Area {
	return &Area{}
}

type Area struct {
	Tiles []*Tile
	Links []Link
}

func (me *Area) Location(p At) (*Tile, error) {
	if p.T > me.Size() {
		return nil, fmt.Errorf("Location %v: tile unknown", p.String())
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

	// set x, y
	for i := 0; i < len(v); i++ {
		from := me.Tiles[v[i].From]
		to := me.Tiles[v[i].To]

		switch v[i].Direction {
		case North:
			to.y = from.y - 1
		case East:
			to.x = from.x + 1
		case South:
			to.y = from.y + 1
		case West:
			to.x = from.x - 1
		}
	}
	// make all x, y positive
	dx := me.minX()
	dy := me.minY()
	for i := 0; i < len(me.Tiles); i++ {
		me.Tiles[i].x -= dx
		me.Tiles[i].y -= dy
	}
	return nil
}

func (me *Area) minX() int {
	var min int
	for i := 0; i < len(me.Tiles); i++ {
		if x := me.Tiles[i].x; x < min {
			min = x
		}
	}
	return min
}

func (me *Area) minY() int {
	var min int
	for i := 0; i < len(me.Tiles); i++ {
		if y := me.Tiles[i].y; y < min {
			min = y
		}
	}
	return min
}

func (me *Area) Size() int { return len(me.Tiles) }

func NewTile() *Tile {
	return &Tile{}
}

type Tile struct {
	Short []byte
	Long  []byte

	x, y int // set by SetLinks
}

func (me *Tile) String() string {
	return fmt.Sprintf("tile %v,%v", me.x, me.y)
}

type Link struct {
	From, To int
	Direction
}

func (me *Link) String() string {
	return fmt.Sprintf("%d %v %d", me.From, me.Direction, me.To)
}

type byXY []*Tile

func (me byXY) Len() int           { return len(me) }
func (me byXY) Swap(i, j int)      { me[i], me[j] = me[j], me[i] }
func (me byXY) Less(i, j int) bool { return me[i].y < me[j].y || me[i].x < me[j].x }