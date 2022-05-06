package cible

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/gregoryv/logger"
	"github.com/gregoryv/nexus"
	"github.com/gregoryv/vt100"
)

func NewUI() *UI {
	return &UI{
		Logger: logger.Silent,

		IO:          NewRWCache(NewStdIO()),
		playerInput: make(chan string, 1),

		out: make(chan Message, 1),
		in:  make(chan Message, 1),
	}
}

type UI struct {
	logger.Logger
	// cache last input/output to simplify tests
	IO          *RWCache
	playerInput chan string

	out chan Message
	in  chan Message

	*Character
	*Tile
}

func (me *UI) Use(c *Client) {
	close(me.out)
	close(me.in)
	me.out = c.Out
	me.in = c.In
}

func (u *UI) Run(ctx context.Context) error {
	u.ShowIntro()

	send := u.out
	e := &EventJoin{Player: Player{
		Name: "x",
	}}
	send <- NewMessage(e)

	// signal when prompt needs update
	promptUpdate := make(chan struct{}, 1)
	go func() {
		t := time.AfterFunc(200*time.Millisecond,
			u.WritePrompt,
		)
		for {
			select {
			case <-ctx.Done():
				return
			case <-promptUpdate:
				t.Stop()
				t = time.AfterFunc(100*time.Millisecond,
					u.WritePrompt,
				)
			}
		}
	}()

	go func() {
		scanner := bufio.NewScanner(u.IO)
		for scanner.Scan() {
			u.playerInput <- scanner.Text()
		}
	}()

	p, _ := nexus.NewPrinter(u)
	// handle incoming messages
	for {
		select {
		case <-ctx.Done():
			return nil
		case m := <-u.in:
			e, known := NewNamedEvent(m.EventName)
			if !known {
				continue
			}
			Decode(e, &m)
			u.handleEvent(e)
			promptUpdate <- struct{}{}

		case input := <-u.playerInput:
			cid := u.CID()
			switch input {
			case "":
			case "n", "w", "s", "e":
				mv := MoveCharacter(cid, nav[input])
				send <- NewMessage(mv)

			case "l": // use a look event
				if u.Tile != nil {
					p.Println(u.Tile.Long)
				}
			case "h", "help":
				u.Write(usage)
			case "q":
				send <- NewMessage(Leave(cid))
				<-time.After(40 * time.Millisecond)
				p.Println("\nBye!")
				return nil
			default:
				e := &EventSay{Ident: cid, Text: input}
				send <- NewMessage(e)
			}
			promptUpdate <- struct{}{}
		}
	}
}

func (u *UI) handleEvent(e interface{}) {
	switch e := e.(type) {

	case CharacterJoin:
		u.OtherPlayer(e.Ident, "joined")

	case EventJoin:
		u.Character = e.Character
		// u.Tile = e.Tile // todo

	case Movement:
		if u.Character.Position.Equal(e.Position) {
			u.Println("cannot move in that direction")
			return
		}
		u.Character.Position = e.Position
		u.Tile = e.Tile
		u.Println(e.Tile.Short)

	case interface{ AffectUI(*UI) }:
		e.AffectUI(u)

	default:
		u.Println("\n", "unknown event: ", e)
	}
}

func (me *UI) CID() Ident {
	if me.Character == nil {
		return ""
	}
	return me.Character.Ident
}

func (me *UI) WritePrompt() {
	fmt.Fprintf(me.IO, "%s> ", me.CID())
}

func (me *UI) Write(p []byte) (int, error) {
	return me.IO.Write(p)
}

func (me *UI) Println(v ...interface{}) (int, error) {
	p, err := nexus.NewPrinter(me.IO)
	p.Println(v...)
	return int(p.Written), *err
}

func (u *UI) ShowIntro() {
	u.Write([]byte("\033c"))
	u.Write(logo)
}

func (u *UI) Do(v string) {
	u.DoWait(v, "")
}

func (u *UI) DoWait(v, duration string) {
	u.playerInput <- v
	dur, err := time.ParseDuration(duration)
	if err != nil {
		dur = 20 * time.Millisecond
	}
	<-time.After(dur)
}

// only for speach
func (me *UI) OtherPlayerSays(id Ident, text string) {
	fmt.Fprintf(me.IO, "\n%s%s: %s%s\n", cyan, id, text, reset)
}

// for notifications
func (me *UI) OtherPlayer(id Ident, text string) {
	fmt.Fprintf(me.IO, "\n%s%s: %s%s\n", yellow, id, text, reset)
}

// ----------------------------------------

type IO io.ReadWriter

func NewRWCache(rw io.ReadWriter) *RWCache {
	return &RWCache{
		ReadWriter: rw,
	}
}

type RWCache struct {
	io.ReadWriter
	LastRead  []byte
	LastWrite []byte
}

func (me *RWCache) Read(p []byte) (int, error) {
	n, err := me.ReadWriter.Read(p)
	me.LastRead = p
	return n, err
}

func (me *RWCache) Write(p []byte) (int, error) {
	n, err := me.ReadWriter.Write(p)
	me.LastWrite = p
	return n, err
}

// ----------------------------------------

func NewBufIO() *BufIO {
	return &BufIO{
		input:  &bytes.Buffer{},
		output: &bytes.Buffer{},
	}
}

type BufIO struct {
	input  *bytes.Buffer
	output *bytes.Buffer
}

func (me *BufIO) Read(p []byte) (int, error) {
	return me.input.Read(p)
}

func (me *BufIO) Write(p []byte) (int, error) {
	return me.output.Write(p)
}

func NewStdIO() *StdIO {
	return &StdIO{
		Reader: os.Stdin,
		Writer: os.Stdout,
	}
}

type StdIO struct {
	io.Reader // input
	io.Writer // output
}

var (
	fg     = vt100.ForegroundColors()
	yellow = fg.Yellow.Bytes()
	cyan   = fg.Cyan.Bytes()

	vt    = vt100.Attributes()
	reset = vt.Reset.Bytes()
)

var nav = map[string]Direction{
	"n": N,
	"e": E,
	"s": S,
	"w": W,
}

var usage = []byte(`
Navigation

n - north
e - east
s - south
w - west

l - look around
q - quit
h, help - for this help
`)

var logo = []byte(`
  ____ ___ ____  _     _____ 
 / ___|_ _| __ )| |   | ____|
| |    | ||  _ \| |   |  _|  
| |___ | || |_) | |___| |___ 
 \____|___|____/|_____|_____|
                             
`)
