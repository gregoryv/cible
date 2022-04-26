package cible

import (
	"testing"
)

func Test_cave(t *testing.T) {
	area := myCave()
	for _, tile := range area.Tiles {
		t.Error(tile, tile.Nav)
	}
}

func myCave() *Area {
	t1 := &Tile{
		Ident: "01",
		Short: "entrance",
		Long:  "longer description",
		Nav:   Nav{N: "02"},
	}

	t2 := &Tile{
		Ident: "02",
		Short: "room with high ceiling",
		Nav:   Nav{E: "03", S: "01"},
	}

	t3 := &Tile{
		Ident: "03",
		Short: "small area",
		Nav:   Nav{W: "02"},
	}

	return &Area{
		Tiles: Tiles{t1, t2, t3},
	}
}
