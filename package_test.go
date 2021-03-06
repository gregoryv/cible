package cible_test

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"testing"
	"time"

	. "github.com/gregoryv/cible"
	. "github.com/gregoryv/cible/tui" // fix this, don't rely on tui to test this package
	"github.com/gregoryv/logger"
)

func TestServer(t *testing.T) {
	srv := NewServer()
	srv.Logger = t
	ctx, cancel := context.WithCancel(context.Background())
	// start server
	go srv.Run(ctx, startNewGame(t))
	pause("10ms")

	red := newUI(t, srv)
	go red.Run(ctx)

	blue := newUI(t, srv)
	go blue.Run(ctx)

	// let clients connect
	<-time.After(100 * time.Millisecond)

	// GAME PLAY
	red.Do("n")         // move north
	red.Do("")          // say nothing
	red.Do("HellOOO!!") // speak
	red.Do("l")         // look around
	red.Do("h")         // help
	blue.DoWait("q", "100ms")
	cancel()
	<-time.After(100 * time.Millisecond)
}

func newUI(t *testing.T, srv *Server) *UI {
	c := NewClient()
	c.Logger = t
	c.Host = srv.Addr().String()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	c.Connect(ctx)
	ui := NewUI()
	ui.Use(c)
	ui.IO = NewRWCache(NewBufIO())
	return ui
}

func TestServer_Run(t *testing.T) {

	t.Run("backoff", func(t *testing.T) {
		srv := NewServer()
		srv.Listener = &brokenListener{}

		var buf bytes.Buffer
		srv.Logger = logger.Wrap(log.New(&buf, "", log.LstdFlags))

		dur := 10 * time.Millisecond
		ctx, _ := context.WithTimeout(context.Background(), dur)
		if err := srv.Run(ctx, nil); err != nil {
			t.Fatal(err)
		}

		errCount := bytes.Count(buf.Bytes(), []byte("broken"))
		if errCount >= 2 {
			t.Errorf("calm down, %v accept failures in %v", errCount, dur)
		}
	})

	t.Run("respect MaxAcceptErrors", func(t *testing.T) {
		srv := NewServer()
		srv.Listener = &brokenListener{}
		srv.MaxAcceptErrors = 1

		var buf bytes.Buffer
		srv.Logger = logger.Wrap(log.New(&buf, "", log.LstdFlags))

		dur := 100 * time.Millisecond
		ctx, _ := context.WithTimeout(context.Background(), dur)
		if err := srv.Run(ctx, nil); err == nil {
			t.Fatal(err)
		}
	})

	t.Run("Bind", func(t *testing.T) {
		srv := NewServer()
		srv.Bind = "jibberish"

		dur := 10 * time.Millisecond
		ctx, _ := context.WithTimeout(context.Background(), dur)
		if err := srv.Run(ctx, nil); err == nil {
			t.Fatal(err)
		}
	})
}

func Test_badEvents(t *testing.T) {
	g := startNewGame(t)
	c := &EventJoinGame{Player: Player{Name: "John"}}
	if err := g.Do(c); err != nil {
		t.Fatal(err)
	}
	cases := []Event{
		&EventMove{Direction: N},               // no such character
		&EventLeave{Ident: "Eve"},              // no such character
		&EventMove{Ident: "god", Direction: N}, // cannot be move)
		&EventMove{Ident: c.Ident, Direction: Direction(-1)},
		&badEvent{err: broken},
	}
	for _, c := range cases {
		t.Run("", func(t *testing.T) {
			if err := g.Do(c); err == nil {
				t.Errorf("%v worked?!", c)
			}
		})
	}
	g.Do(&EventStopGame{})
	if err := g.Do(&EventMove{Ident: c.Ident, Direction: N}); err == nil {
		t.Error("games is stopped but event was done")
	}

}

func Test_cancelGame(t *testing.T) {
	g := NewGame()
	ctx, cancel := context.WithCancel(context.Background())
	go g.Run(ctx)
	cancel()
}

func Test_cave(t *testing.T) {
	for _, tile := range Spaceport().Tiles {
		t.Log(tile, tile.Nav)
	}
}

func TestArea_Tile(t *testing.T) {
	var a Area
	if _, err := a.Tile("x"); err == nil {
		t.Fail()
	}
}

func TestDirection(t *testing.T) {
	_ = Direction(-1).String() // should work
}

// ----------------------------------------

func BenchmarkMoveCharacter_1_player(b *testing.B) {
	g := startNewGame(b)
	defer g.Do(&EventStopGame{})

	e := &EventJoinGame{Player: Player{Name: "John"}}
	g.Do(e)
	cid := e.Ident
	for i := 0; i < b.N; i++ {
		g.Do(&EventMove{Ident: cid, Direction: N})
		g.Do(&EventMove{Ident: cid, Direction: S})
	}
}

func BenchmarkMoveCharacter_1000_player(b *testing.B) {
	g := startNewGame(b)
	defer g.Do(&EventStopGame{})

	// Join all players first
	for i := 0; i < 1000; i++ {
		var p Player
		p.SetName(fmt.Sprintf("John%v", i))
		e := &EventJoinGame{Player: p}
		if err := g.Do(e); err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cid := Ident(fmt.Sprintf("John%v", rand.Intn(1000)))
		g.Do(&EventMove{Ident: cid, Direction: N})
		g.Do(&EventMove{Ident: cid, Direction: S})
	}
}

// ----------------------------------------

type badEvent struct {
	err error
}

func (e *badEvent) Event() string          { return "badEvent" }
func (e *badEvent) Done() error            { return e.err }
func (e *badEvent) AffectGame(*Game) error { return e.err }

func startNewGame(t testing.TB) *Game {
	g := NewGame()
	g.Logger = t
	g.LogAllEvents = true
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(func() {
		g.Logger = logger.Silent
		cancel()
	})
	go g.Run(ctx)
	time.Sleep(10 * time.Millisecond) // let it start
	return g
}

func pause(v string) {
	dur, err := time.ParseDuration(v)
	if err != nil {
		panic(err.Error())
	}
	<-time.After(dur)
}

type brokenListener struct{}

func (me *brokenListener) Accept() (net.Conn, error) { return nil, broken }
func (me *brokenListener) Addr() net.Addr            { return nil }
func (me *brokenListener) Close() error              { return broken }

var broken = fmt.Errorf("broken")
