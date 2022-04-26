package cible

import (
	"testing"
)

func Test_cave(t *testing.T) {
	area := myCave()
	for _, tile := range area {
		t.Error(tile, tile.Links())
	}
}

func myCave() []*Tile {
	t1 := &Tile{
		id:    "01",
		short: "entrance",
		long:  "longer description",
		links: Nav{N: "02"},
	}

	t2 := &Tile{
		id:    "02",
		short: "room with high ceiling",
		links: Nav{E: "03", S: "01"},
	}

	t3 := &Tile{
		id:    "03",
		short: "small area",
		links: Nav{W: "02"},
	}

	return []*Tile{t1, t2, t3}
}
