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
