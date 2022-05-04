package cible

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"testing"
	"time"

	"github.com/gregoryv/logger"
)

func xTestServer(t *testing.T) {
	g := startNewGame(t)
	srv := NewServer()
	srv.Logger = t
	// so we don't log After test is done
	defer func() { srv.Logger = logger.Silent }()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	// red bot
	red := NewClient()
	red.Logger = t

	// connect if server is down, should not work
	if err := red.Connect(ctx); err == nil {
		t.Error("connected to nothing?")
	}

	// start server
	go srv.Run(ctx, g)
	pause("10ms")

	red.Host = srv.Addr().String()
	redui := NewUI()
	go redui.Run(ctx)

	// blue bot
	blue := NewClient()
	blue.Logger = t
	blue.Host = srv.Addr().String()
	blueui := NewUI()
	go blueui.Run(ctx)

	// GAME PLAY
	// this does get us coverage but it doesn't verify anything
	// let them connect
	<-time.After(200 * time.Millisecond)
	// move
	redui.Do("n")

	redui.Do("")          // say nothing
	redui.Do("HellOOO!!") // speak
	redui.Do("l")         // look around
	// try to hack
	red.Out <- NewMessage(&badEvent{})

	redui.Do("h") // help
	redui.Do("q") // leave game
	//blueui.DoWait("q", "200ms")
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
	c := Join(Player{Name: "John"})

	if err := g.Do(c); err != nil {
		t.Fatal(err)
	}
	cases := []GameEvent{
		MoveCharacter("Eve", N), // no such playe)
		Leave("Eve"),            // no such playe)
		MoveCharacter("god", N), // cannot be move)
		MoveCharacter(c.Ident, Direction(-1)),
		&badEvent{err: broken},
	}
	for _, c := range cases {
		t.Run("", func(t *testing.T) {
			if err := g.Do(c); err == nil {
				t.Errorf("%v worked?!", c)
			}
		})
	}
	g.Do(StopGame())
	if err := g.Do(MoveCharacter(c.Ident, N)); err == nil {
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
	for _, tile := range myCave().Tiles {
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
	defer g.Do(StopGame())

	e := Join(Player{Name: "John"})
	g.Do(e)
	cid := e.Ident
	for i := 0; i < b.N; i++ {
		g.Do(MoveCharacter(cid, N))
		g.Do(MoveCharacter(cid, S))
	}
}

func BenchmarkMoveCharacter_1000_player(b *testing.B) {
	g := startNewGame(b)
	defer g.Do(StopGame())

	// Join all players first
	for i := 0; i < 1000; i++ {
		p := Player{Name: Name(fmt.Sprintf("John%v", i))}
		e := Join(p)
		if err := g.Do(e); err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cid := Ident(fmt.Sprintf("John%v", rand.Intn(1000)))
		g.Do(MoveCharacter(cid, N))
		g.Do(MoveCharacter(cid, S))
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
