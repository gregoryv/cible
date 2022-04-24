package cible

func NewWorld() *World {
	return &World{
		atlas: make(map[int]Location),
	}
}

type World struct {
	atlas map[int]Location
}

type Location struct {
	id int // same as atlas[id]
}
