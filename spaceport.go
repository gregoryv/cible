package cible

func Spaceport() *Area {
	t1 := &Tile{
		Short: "Center Stateroom",
		Long:  `...`,
	}

	t2 := &Tile{
		Short: "South Stateroom",
		Long:  `...`,
	}

	t3 := &Tile{
		Short: "South-west Stateroom",
		Long:  `...`,
	}

	t4 := &Tile{
		Short: "West Stateroom",
		Long:  `...`,
	}

	t5 := &Tile{
		Short: "North-west Stateroom",
		Long:  `...`,
	}

	t6 := &Tile{
		Short: "North Stateroom",
		Long:  `...`,
	}

	t7 := &Tile{
		Short: "North-east Stateroom",
		Long:  `...`,
	}

	t8 := &Tile{
		Short: "East Stateroom",
		Long:  `...`,
	}

	t9 := &Tile{
		Short: "South-east Stateroom",
		Long:  `...`,
	}

	area := &Area{
		Ident: "a1",
		Title: "Spaceport",
	}
	area.AddTile(t1, t2, t3, t4, t5, t6, t7, t8, t9)

	// link only when they have ids, which is assigned by Area.AddTile
	t1.Link(
		t2, S,
		t3, SW,
		t4, W,
		t5, NW,
		t6, N,
	)
	t2.Link(t3, W)
	t3.Link(t4, N)
	t4.Link(t5, N)
	t5.Link(t6, E)
	t6.Link(t7, E)
	t7.Link(t8, S)
	t8.Link(t9, S)
	return area
}
