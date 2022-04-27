package cible

func Join(p Player) *EventJoin {
	return &EventJoin{
		Player: p,
	}
}

type EventJoin struct {
	Player

	Ident // set when done

}

func (e *EventJoin) Affect(g *Game) error {
	g.Logf("%s join", e.Player.Name)
	p := Position{
		Area: "a1", Tile: "01",
	}
	c := &Character{
		Ident:    Ident(e.Player.Name),
		Name:     e.Player.Name,
		Position: p,
	}
	g.Characters = append(g.Characters, c)
	e.Ident = c.Ident
	return nil
}

// ----------------------------------------

func Leave(cid Ident) *EventLeave {
	return &EventLeave{
		Ident: cid,
	}
}

type EventLeave struct {
	Ident
}

func (e *EventLeave) Affect(g *Game) error {
	c, err := g.Character(e.Ident)
	if err != nil {
		return err
	}
	g.Logf("%s left", c.Name)
	return nil
}
