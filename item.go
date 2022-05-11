package cible

type Items []Item

type Item struct {
	Name
	Count uint

	Position // if it's not in a persons inventory
}
