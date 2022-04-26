package cible

//go:generate stringer -type Direction direction.go
type Direction int

const (
	N Direction = iota
	E
	S
	W
)

func (me Direction) Opposite() Direction {
	switch me {
	case N:
		return S
	case E:
		return W
	case W:
		return E
	}
	return me
}
