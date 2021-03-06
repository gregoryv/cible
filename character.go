package cible

import "strings"

type Character struct {
	Ident
	Name
	Location
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
	nearby := g.Characters.At(me.Location)
	for _, c := range nearby {
		if c.Ident == me.Ident {
			continue
		}
		g.Logf("transmit %s to %s", m.String(), c.Ident)
		c.Transmit(m)
	}
	return nil
}

type Player struct {
	Name
}

type Bot struct{}

type Location struct {
	Area Ident
	Tile Ident
}

func (p *Location) Equal(v Location) bool {
	return p.Area == v.Area && p.Tile == v.Tile
}

func NewInventory() *Inventory {
	return &Inventory{
		Items: Items{
			&Item{
				Name:  "credit",
				Count: 200,
			},
			&Item{
				Name:  "Communicator",
				Count: 1,
			},
			&Item{
				Name:  "Digipass",
				Count: 1,
			},
		},
	}
}

type Inventory struct {
	Items
}

func (me *Inventory) AddItem(v Item) {
	v.Name = Name(strings.Title(string(v.Name)))
	for i, item := range me.Items {
		if item.Name == v.Name {
			me.Items[i].Count++
			return
		}
	}

	me.Items = append(me.Items, &v)
}
