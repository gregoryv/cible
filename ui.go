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
	"github.com/gregoryv/vt100"
)

func NewUI() *UI {
	return &UI{
		Logger: logger.Silent,
		IO:     NewRWCache(NewStdIO()),
		out:    make(chan Message, 1),
		in:     make(chan Message, 1),
	}
}

type UI struct {
	logger.Logger
	out chan Message
	in  chan Message

	// cache last input/output to simplify tests
	IO *RWCache
}

func (me *UI) Use(c *Client) {
	close(me.out)
	close(me.in)
	me.out = c.Out
	me.in = c.In
}

func (u *UI) Run(ctx context.Context) error {
	u.ShowIntro()

	// create player and join game
	p := Player{Name: Name(os.Getenv("USER"))}
	j := &EventJoin{Player: p}

	send := u.out
	send <- NewMessage(j)

	msg := <-u.in
	if err := Decode(j, &msg); err != nil {
		return err
	}

	cid := j.Ident
	// uggly way to set current pos, todo fix it
	m := MoveCharacter(cid, N)
	send <- NewMessage(m)
	m.Direction = S
	send <- NewMessage(m)

	writePrompt := func() { fmt.Fprintf(u, "%s> ", cid) }
	playerInput := make(chan string, 1)

	scanErr := make(chan error)
	go func() {
		scanner := bufio.NewScanner(u.IO)
		writePrompt()
		for scanner.Scan() {
			playerInput <- scanner.Text()
			writePrompt()
		}
		scanErr <- scanner.Err()
	}()

	// handle incoming messages
	for {
		select {
		case err := <-scanErr:
			return err
		case <-ctx.Done():
			return nil
		case m := <-u.in:
			e, known := NewNamedEvent(m.EventName)
			if !known {
				continue
			}
			Decode(e, &m)
			if e, ok := e.(interface{ AffectUI(*UI) }); ok {
				e.AffectUI(u)
			} else {
				fmt.Fprintf(u, "\n%s%v%s\n", yellow, e, reset)
			}
			writePrompt()

		case input := <-playerInput:
			switch input {
			case "":
				// ignore
			case "n", "w", "s", "e":
				mv := MoveCharacter(cid, nav[input])
				send <- NewMessage(mv)

			case "l":
				// todo first position
				if m.Tile != nil {
					u.Write([]byte(m.Tile.Long))
					fmt.Println()
				}
			case "h", "help":
				u.Write(usage)
			case "q":
				send <- NewMessage(Leave(cid))
				<-time.After(40 * time.Millisecond)
				fmt.Fprintln(u, "\nBye!")
				return nil
			default:
				e := &EventSay{Ident: cid, Text: input}
				send <- NewMessage(e)
			}
		}
	}
}

func (me *UI) Write(p []byte) (int, error) {
	return me.IO.Write(p)
}

func (u *UI) ShowIntro() {
	u.Write([]byte("\033c"))
	u.Write(logo)
}

func (me *UI) Do(v string) {
	me.DoWait(v, "")
}

func (me *UI) DoWait(v, duration string) {
	me.IO.Write([]byte(v))
	me.IO.Write([]byte("\n"))
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
