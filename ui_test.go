package cible

import (
	"context"
	"strings"
	"testing"
)

func TestUI_Run(t *testing.T) {
	ui := NewUI()
	if err := ui.Run(context.Background(), NewClient()); err == nil {
		t.Error("expected error on client failure")
	}
}

func TestUI_OtherPlayerSays(t *testing.T) {
	ui := NewUI()

	ui.OtherPlayerSays("cid", "hello")
	got, exp := string(ui.IO.LastWrite), "hello"
	if !strings.Contains(got, exp) {
		t.Errorf("%s\nmissing %s", got, exp)
	}
}

func TestUI_OtherPlayer(t *testing.T) {
	ui := NewUI()

	ui.OtherPlayer("cid", "left")
	got, exp := string(ui.IO.LastWrite), "left"
	if !strings.Contains(got, exp) {
		t.Errorf("%s\nmissing %s", got, exp)
	}
}
