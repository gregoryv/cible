package cible

//go:generate stringer -type Direction direction.go
type Direction int

const (
	North Direction = iota
	NorthEast
	East
	SouthEast
	South
	SouthWest
	West
	NorthWest
)

func (me Direction) Opposite() Direction {
	switch me {
	case North:
		return South
	case NorthEast:
		return SouthWest
	case East:
		return West
	case SouthEast:
		return NorthWest
	case West:
		return East
	case NorthWest:
		return SouthEast
	}
	return me
}
