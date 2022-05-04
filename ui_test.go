package cible

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestUI_Run(t *testing.T) {
	ui := NewUI()
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	go ui.Run(ctx)
	select {
	case <-ctx.Done():
		t.Fail()
	case m := <-ui.out:
		// got first join message
		if m.EventName != "cible.EventJoin" {
			t.Error("expected EventJoin got: ", m.String())
		}
	}
	cancel()
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
