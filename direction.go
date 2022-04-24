package cible

//go:generate stringer -type Direction direction.go
type Direction int

const (
	DirectionUnknown Direction = iota
	Forward
	Backward
	Left
	Right
)
