package cible

// fields are short types, with optional simple set methods

type Name string

func (me *Name) SetName(v string) { *me = Name(v) }

type Short string
type Long string
type Title string
type IsBot bool

type Ident string

func (me *Ident) SetIdent(v string) { *me = Ident(v) }
