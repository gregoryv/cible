package cible

import (
	"bytes"
	"strings"
	"testing"
)

func TestUI_OtherPlayerSays(t *testing.T) {
	ui := NewUI()
	var buf bytes.Buffer
	ui.stdout = &buf

	ui.OtherPlayerSays("cid", "hello")
	got, exp := buf.String(), "hello"
	if !strings.Contains(got, exp) {
		t.Errorf("%s\nmissing %s", got, exp)
	}
}

func TestUI_OtherPlayer(t *testing.T) {
	ui := NewUI()
	var buf bytes.Buffer
	ui.stdout = &buf

	ui.OtherPlayer("cid", "left")
	got, exp := buf.String(), "left"
	if !strings.Contains(got, exp) {
		t.Errorf("%s\nmissing %s", got, exp)
	}
}
