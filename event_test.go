package cible

import (
	"strings"
	"testing"
)

func TestEventMove(t *testing.T) {
	var p Player
	p.SetName("John")
	e := &EventMove{
		player: p,
		dir:    S,
	}
	got := e.Event()
	if !strings.Contains(got, "John") {
		t.Errorf("missing name: %q", got)
	}
}
