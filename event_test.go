package cible

import (
	"strings"
	"testing"
)

func TestEventMove(t *testing.T) {
	e := &EventMove{
		player: Player{
			name: "john",
		},
		dir: South,
	}
	got := e.Event()
	if !strings.Contains(got, "john") {
		t.Errorf("missing name: %q", got)
	}
}
