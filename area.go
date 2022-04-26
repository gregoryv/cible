package cible

import "fmt"

func NewArea() *Area {
	return &Area{}
}

type Area struct {
	Tiles []*Tile
	Links []Link
}

func (me *Area) Location(p At) (*Tile, error) {
	if p.T >= me.Size() {
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
