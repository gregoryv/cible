package cible

import (
	"fmt"
	"strings"
)

type Tile struct {
	id    Ident
	short string
	long  string

	links Nav // the four directions
}

func (me *Tile) SetId(v Ident)     { me.id = v }
func (me *Tile) SetShort(v string) { me.short = v }
func (me *Tile) SetLong(v string)  { me.long = v }
func (me *Tile) SetLinks(v Nav)    { me.links = v }

func (me *Tile) Links() Nav { return me.links }

func (me *Tile) String() string {
	return fmt.Sprintf("%s %s", me.id, me.short)
}

type Nav [4]Ident

func (me Nav) String() string {
	var res []string
	for d, id := range me {
		if id != "" {
			res = append(res, Direction(d).String()+":"+string(id))
		}
	}
	return strings.Join(res, " ")
}

type Ident string
