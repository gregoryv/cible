package cible

type Character struct {
	Ident
	Name
	Position
	IsBot
	Inventory

	tr Transmitter // set by server for communication
}

func (me *Character) Transmit(m Message) error {
	if me.tr == nil { // ie. if bot
		return nil
	}
	return me.tr.Transmit(m)
}

func (me *Character) TransmitOthers(g *Game, m Message) error {
	nearby := g.Characters.At(me.Position)
	for _, c := range nearby {
		if c.Ident == me.Ident {
			continue
		}
		g.Logf("transmit %s to %s", m.String(), c.Ident)
		c.Transmit(m)
	}
	return nil
}

type IsBot bool

type Player struct {
	Name
}

type Bot struct{}

type Position struct {
	Area Ident
	Tile Ident
}

func (p *Position) Equal(v Position) bool {
	return p.Area == v.Area && p.Tile == v.Tile
}

type Ident string

func (me *Ident) SetIdent(v string) { *me = Ident(v) }

type Inventory struct {
	Items
}

func (me *Inventory) AddItem(v Item) {
	me.Items = append(me.Items, v)
}

type Items []Item

type Item struct {
	Name
	Count uint
}
