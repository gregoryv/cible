package cible

//go:generate stringer -type Direction direction.go
type Direction int

const (
	North Direction = iota
	East
	South
	West
)

func (me Direction) Opposite() Direction {
	switch me {
	case North:
		return South
	case East:
		return West
	case West:
		return East
	}
	return me
}
