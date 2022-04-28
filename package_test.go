package cible

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/gregoryv/logger"
)

func TestServer(t *testing.T) {
	g := startNewGame(t)
	srv := NewServer()
	srv.Logger = t
	// so we don't log After test is done
	defer func() { srv.Logger = logger.Silent }()

	// start server
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go srv.Run(ctx, g)
	pause("10ms")

	// connect client
	c := NewClient()
	c.Logger = t
	c.Host = srv.Addr.String()
	_ = c.Connect(ctx)

	p := Player{Name: "test"}
	j, err := Send(c, &EventJoin{Player: p})
	if j.Ident == "" {
		t.Error("join failed, missing ident", err)
	}
	t.Log(j)
	m, err := Send(c, MoveCharacter(j.Ident, N))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(m)

}

func TestGame_play(t *testing.T) {
	g := startNewGame(t)
	g.Logger = t

	p := Player{Name: "John"}
	c, j := Trigger(g, Join(p))
	if err := c.Done(); err != nil {
		t.Fatal(err)
	}
	TriggerWait(g, MoveCharacter(j.Ident, N))
	TriggerWait(g, MoveCharacter(j.Ident, E))
	task, e := Trigger(g, MoveCharacter(j.Ident, W))
	task.Done()
	pos := e.Position
	if pos.Tile != "02" {
		t.Error("got", pos.Tile, "exp", "02")
	}
	_, tile, err := g.Place(pos)
	if err != nil {
		t.Fatal(tile, err)
	}
	task, _ = Trigger(g, Leave(j.Ident))
	if err := task.Done(); err != nil {
		t.Error(err)
	}
}

func Test_badEvents(t *testing.T) {
	g := startNewGame(t)

	p := Player{Name: "John"}
	task, c := Trigger(g, Join(p))
	if err := task.Done(); err != nil {
		t.Fatal(err)
	}
	cid := c.Ident
	Trigger(g, MoveCharacter("Eve", N)) // no such playe)
	Trigger(g, MoveCharacter("god", N)) // cannot be move)
	Trigger(g, MoveCharacter(cid, Direction(-1)))
	Trigger(g, &badEvent{})
	TriggerWait(g, MoveCharacter(cid, W))
	task, _ = Trigger(g, Leave("no such"))
	if err := task.Done(); err == nil {
		t.Error("Leave unknown cid should fail")
	}
}

func TestEvent_Done(t *testing.T) {
	// can't use startNewGame here as
	g := NewGame()
	ctx, cancel := context.WithCancel(context.Background())
	go g.Run(ctx)
	t.Cleanup(cancel)

	pause("10ms")
	// stopped in last subtest
	t.Run("Join", func(t *testing.T) {
		e := TriggerWait(g, Join(Player{Name: "John"}))
		first := e.Ident
		if first != e.Ident {
			t.Error("multiple calls to Done gave different values")
		}
	})

	t.Run("MoveCharacter", func(t *testing.T) {
		task, e := Trigger(g, MoveCharacter("John", N))
		task.Done()
		first := e.Position
		task.Done()
		if !first.Equal(e.Position) {
			t.Error("multiple calls to Done gave different values")
		}
	})

	t.Run("Leave", func(t *testing.T) {
		task, _ := Trigger(g, Leave("no such"))
		task.Done()
		task.Done()
	})

	// keep last as it stops game
	t.Run("StopGame", func(t *testing.T) {
		task, _ := Trigger(g, StopGame())
		task.Done()
		task.Done()
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
	TriggerWait(g, StopGame())

	for _, e := range events {
		t.Run(fmt.Sprintf("%T", e), func(t *testing.T) {
			TriggerWait(g, e)
		})
	}
}

type badEvent struct {
	err error
}

func (e *badEvent) Event() string      { return "badEvent" }
func (e *badEvent) Done() error        { return e.err }
func (e *badEvent) Affect(*Game) error { return e.err }

func startNewGame(t *testing.T) *Game {
	g := NewGame()
	g.Logger = t
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

// ----------------------------------------

func BenchmarkMoveCharacter_1_player(b *testing.B) {
	g := NewGame()
	go g.Run(context.Background())
	defer Trigger(g, StopGame())

	p := Player{Name: "John"}
	task, e := Trigger(g, Join(p))
	if err := task.Done(); err != nil {
		b.Fatal(err)
	}
	cid := e.Ident
	for i := 0; i < b.N; i++ {
		TriggerWait(g, MoveCharacter(cid, N))
		TriggerWait(g, MoveCharacter(cid, S))
	}
}

func BenchmarkMoveCharacter_1000_player(b *testing.B) {
	g := NewGame()
	go g.Run(context.Background())
	defer Trigger(g, StopGame())

	for i := 0; i < 1000; i++ {
		p := Player{Name: Name(fmt.Sprintf("John%v", i))}
		task, _ := Trigger(g, Join(p))
		if err := task.Done(); err != nil {
			b.Fatal(err)
		}
	}

	for i := 0; i < b.N; i++ {
		cid := Ident(fmt.Sprintf("John%v", rand.Intn(1000)))
		TriggerWait(g, MoveCharacter(cid, N))
		TriggerWait(g, MoveCharacter(cid, S))
	}
}
