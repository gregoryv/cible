package tui

import (
	"strings"
	"testing"
)

func TestUI_OtherPlayerSays(t *testing.T) {
	tui := NewUI()

	tui.OtherPlayerSays("cid", "hello")
	got, exp := string(tui.IO.LastWrite), "hello"
	if !strings.Contains(got, exp) {
		t.Errorf("%s\nmissing %s", got, exp)
	}
}

func TestUI_OtherPlayer(t *testing.T) {
	tui := NewUI()

	tui.OtherPlayer("cid", "left")
	got, exp := string(tui.IO.LastWrite), "left"
	if !strings.Contains(got, exp) {
		t.Errorf("%s\nmissing %s", got, exp)
	}
}
