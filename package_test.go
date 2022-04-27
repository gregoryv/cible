package cible

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestServer(t *testing.T) {
	g := startNewGame(t)
	srv := NewServer()
	srv.Logger = t
	ctx, cancel := context.WithCancel(context.Background())
	_ = time.AfterFunc(100*time.Millisecond, cancel)
	if err := srv.Run(ctx, g); err != nil {
		t.Error(err)
	}
}

func TestGame_play(t *testing.T) {
	g := startNewGame(t)
	g.Logger = t

	p := Player{Name: "John"}
	c := Trigger(g, Join(p))
	t.Log("here")
	if err := c.Done(); err != nil {
		t.Fatal(err)
	}
	cid := c.Ident
	Trigger(g, MoveCharacter(cid, N)).Done()
	Trigger(g, MoveCharacter(cid, E)).Done()

	e := Trigger(g, MoveCharacter(cid, W))
	e.Done()
	pos := e.Position
	if pos.Tile != "02" {
		t.Error("got", pos.Tile, "exp", "02")
	}
	_, tile, err := g.Place(pos)
	if err != nil {
		t.Fatal(tile, err)
	}
	if err := Trigger(g, Leave(cid)).Done(); err != nil {
		t.Error(err)
	}
}

func Test_badEvents(t *testing.T) {
	g := startNewGame(t)

	p := Player{Name: "John"}
	c := Trigger(g, Join(p))
	if err := c.Done(); err != nil {
		t.Fatal(err)
	}
	cid := c.Ident
	g.Events <- MoveCharacter("Eve", N) // no such playe)
	g.Events <- MoveCharacter("god", N) // cannot be move)
	g.Events <- MoveCharacter(cid, Direction(-1))
	g.Events <- &badEvent{}
	Trigger(g, MoveCharacter(cid, W)).Done()
	e := Trigger(g, Leave("no such"))
	if err := e.Done(); err == nil {
		t.Error("Leave unknown cid should fail")
	}
}

func TestEvent_Done(t *testing.T) {
	// can't use startNewGame here as
	g := NewGame()
	ctx, cancel := context.WithCancel(context.Background())
	go g.Run(ctx)
	t.Cleanup(cancel)

	<-time.After(10 * time.Millisecond)
	// stopped in last subtest
	t.Run("Join", func(t *testing.T) {
		e := Trigger(g, Join(Player{Name: "John"}))
		e.Done()
		first := e.Ident
		e.Done()
		if first != e.Ident {
			t.Error("multiple calls to Done gave different values")
		}
	})

	t.Run("MoveCharacter", func(t *testing.T) {
		e := Trigger(g, MoveCharacter("John", N))
		e.Done()
		first := e.Position
		e.Done()
		if !first.Equal(e.Position) {
			t.Error("multiple calls to Done gave different values")
		}
	})

	t.Run("Leave", func(t *testing.T) {
		e := Trigger(g, Leave("no such"))
		e.Done()
		e.Done()
	})

	// keep last as it stops game
	t.Run("StopGame", func(t *testing.T) {
		e := Trigger(g, StopGame())
		e.Done()
		e.Done()
	})
}

func catchPanic(t *testing.T) {
	if err := recover(); err != nil {
		t.Helper()
		t.Fatal(err)
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

func TestTrigger(t *testing.T) {
	g := startNewGame(t) // not running
	events := []Event{
		Join(Player{Name: "John"}),
		MoveCharacter("John", N),
		Leave("John"),
		StopGame(),
	}
	Trigger(g, StopGame()).Done()

	for _, e := range events {
		t.Run(fmt.Sprintf("%T", e), func(t *testing.T) {
			Trigger(g, e).Done()
		})
	}
}

type badEvent struct {
	err error
}

func (me *badEvent) Event() string      { return "badEvent" }
func (me *badEvent) Done() error        { return me.err }
func (me *badEvent) Affect(*Game) error { return me.err }
func (e *badEvent) setErr(v error) {
	e.err = v
}

func startNewGame(t *testing.T) *Game {
	g := NewGame()
	t.Cleanup(func() {
		Trigger(g, StopGame()).Done() // wait for it to complete
	})
	go g.Run(context.Background())
	time.Sleep(10 * time.Millisecond) // let it start
	return g
}

// ----------------------------------------

func BenchmarkMoveCharacter_1_player(b *testing.B) {
	g := NewGame()
	go g.Run(context.Background())
	defer g.Stop()

	p := Player{Name: "John"}
	e := Trigger(g, Join(p))
	if err := e.Done(); err != nil {
		b.Fatal(err)
	}
	cid := e.Ident
	for i := 0; i < b.N; i++ {
		Trigger(g, MoveCharacter(cid, N)).Done()
		Trigger(g, MoveCharacter(cid, S)).Done()
	}
}

func BenchmarkMoveCharacter_1000_player(b *testing.B) {
	g := NewGame()
	go g.Run(context.Background())
	defer g.Stop()

	for i := 0; i < 1000; i++ {
		p := Player{Name: Name(fmt.Sprintf("John%v", i))}
		e := Trigger(g, Join(p))
		if err := e.Done(); err != nil {
			b.Fatal(err)
		}
	}

	for i := 0; i < b.N; i++ {
		cid := Ident(fmt.Sprintf("John%v", rand.Intn(1000)))
		Trigger(g, MoveCharacter(cid, N)).Done()
		Trigger(g, MoveCharacter(cid, S)).Done()
	}
}
