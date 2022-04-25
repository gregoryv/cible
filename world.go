package cible

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
}

func (me *Area) AddTile(t *Tile) {
	t.Id = len(me.tiles)
	me.tiles = append(me.tiles, t)
}

func (me *Area) Size() int { return len(me.tiles) }

func NewTile() *Tile {
	return &Tile{}
}

type Tile struct {
	Id    int
	Short []byte
	Long  []byte
}
