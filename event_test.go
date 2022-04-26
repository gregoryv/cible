package cible

import (
	"strings"
	"testing"
)

func TestEventMove(t *testing.T) {
	p := Player{
		Name: "John",
	}
	e := &EventMove{p, S}
	got := e.Event()
	if !strings.Contains(got, "John") {
		t.Errorf("missing name: %q", got)
	}
}
