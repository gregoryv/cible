package cible

func Spaceport() *Area {
	t1 := &Tile{
		Short: "Center Stateroom",

		Long: `

A large tree with pinkish fruits grows in the center. Surrounded by
benches with soft padding. The large stateroom is bright and the
ceiling transparently shows the galaxy augmented with names of nearest
starsystems. Alpha Centauri, Barnard's Star and Luhman 16 all sparkle
in bright colors.

To the east the great memorial wall with it's soft and rock like
surface, reminds you of the venturesome life in space.`}

	t2 := &Tile{
		Short: "South Stateroom",
		Long:  `Open space`,
	}

	t3 := &Tile{
		Short: "Tech room",

		Long: `

Historical information about the Genetic Low Orbital Computer is
posted on the wall. Tech pillars with charging ports and terminals are
vacant for use.`,
		//
	}

	t4 := &Tile{
		Short: "West Stateroom",
		Long:  `Couple of drink and food dispnesers are humming.`,
	}

	t5 := &Tile{
		Short: "Sitting room",
		Long:  `A lounge with some tables and chairs.`,
	}

	t6 := &Tile{
		Short: "News room",
		Long:  `On the north wall news are displayed on a multi screen setup.`,
	}

	t7 := &Tile{
		Short: "North-east Stateroom",
		Long:  `Open space`,
	}

	t8 := &Tile{
		Short: "Rest room",
		Long:  `Multiple toilets are available, some are occupied or just broken`,
	}

	t9 := &Tile{
		Short: "South-east Stateroom",
		Long:  `Open space`,
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
	t2.Link(
		t3, W,
		t9, E,
	)
	t3.Link(t4, N)
	t4.Link(t5, N)
	t5.Link(t6, E)
	t6.Link(t7, E)
	t7.Link(t8, S)
	t8.Link(t9, S)
	return area
}
