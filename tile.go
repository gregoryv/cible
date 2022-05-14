package cible

import "fmt"

type Tile struct {
	Ident
	Short
	Long
	Nav

	*Cybromat
}

func (t *Tile) String() string {
	return fmt.Sprintf("%s %s", t.Ident, t.Short)
}

// Link creates a dual link between a tile and the given ones
func (me *Tile) Link(to ...interface{}) {
	for i := 0; i < len(to); i += 2 {
		t := to[i].(*Tile)
		d := to[i+1].(Direction)
		if me.Nav[d] != "" {
			if me.Nav[d] == t.Ident {
				continue // already linked
			}
			panic(
				fmt.Sprintf(
					"cannot link %s, %s already linked to %v",
					me.String(), d.String(), me.Nav[d],
				),
			)
		}
		// link in both directions
		me.Nav[d] = t.Ident
		t.Nav[opposite[d]] = me.Ident
	}
}
