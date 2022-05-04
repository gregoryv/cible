package cible

import (
	"context"
	"strings"
	"testing"
)

func TestUI_Run(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	ui := NewUI()
	if err := ui.Run(ctx, NewClient()); err == nil {
		t.Error("expected error on client failure")
	}
	cancel()
}

func TestUI_OtherPlayerSays(t *testing.T) {
	ui := NewUI()
	io := NewBufIO()
	ui.IO = io

	ui.OtherPlayerSays("cid", "hello")
	got, exp := io.output.String(), "hello"
	if !strings.Contains(got, exp) {
		t.Errorf("%s\nmissing %s", got, exp)
	}
}

func TestUI_OtherPlayer(t *testing.T) {
	ui := NewUI()
	io := NewBufIO()
	ui.IO = io

	ui.OtherPlayer("cid", "left")
	got, exp := io.output.String(), "left"
	if !strings.Contains(got, exp) {
		t.Errorf("%s\nmissing %s", got, exp)
	}
}
