package cible

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gregoryv/logger"
	"github.com/gregoryv/nexus"
	"github.com/nathan-fiscaletti/consolesize-go"
)

func NewUI() *UI {
	cols, rows := consolesize.GetConsoleSize()
	if cols <= 0 {
		cols = 72 // default when testing
		rows = 20
	}
	return &UI{
		Logger: logger.Silent,

		IO:          NewRWCache(NewStdIO()),
		playerInput: make(chan string, 1),

		out: make(chan Message, 1),
		in:  make(chan Message, 1),

		cols:       cols,
		rows:       rows,
		TextFormat: &TextFormat{cols: cols},
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

	cols, rows int
	*TextFormat
}

func (me *UI) Use(c *Client) {
	close(me.out)
	close(me.in)
	me.out = c.Out
	me.in = c.In
}

func (u *UI) Run(ctx context.Context) error {
	u.clearScreen()
	u.ShowIntro()

	send := u.out
	player := Player{}
	player.SetName(os.Getenv("USER"))
	e := &PlayerJoin{Player: player}
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
			e, known := NewEvent(m.EventName)
			if !known {
				continue
			}
			Decode(e, &m)
			u.HandleEvent(e)
			promptUpdate <- struct{}{}

		case input := <-u.playerInput:
			switch input {
			case "":
			case "n", "ne", "e", "se", "s", "sw", "w", "nw":
				send <- NewMessage(&EventMove{Direction: nav[input]})

			case "l":
				send <- NewMessage(&EventLook{})

			case "h", "help":
				u.CenterPrintln(u.boxed(usage, 40))

			case "q":
				send <- NewMessage(&EventLeave{})
				<-time.After(40 * time.Millisecond)
				p.Println("\nBye!")
				return nil

			default:
				if input != "" {
					send <- NewMessage(&EventSay{Text: input})
				}
			}
			promptUpdate <- struct{}{}
		}
	}
}

func (u *UI) HandleEvent(e interface{}) {
	switch e := e.(type) {

	case *EventSay:
		u.OtherPlayerSays(e.Ident, e.Text)

	case *EventJoin:
		u.OtherPlayer(e.Ident, "joined")

	case *PlayerJoin:
		u.Character = e.Character

	case *EventLeave:
		u.OtherPlayer(e.Ident, "left game")

	case *EventLook:
		u.showTile(&e.Tile)

	case *EventMove:
		if u.Character.Position.Equal(e.Position) {
			u.Println("cannot move in that direction")
			return
		}
		u.Character.Position = e.Position
		u.Println(string(e.Body))

	default:
		u.Println("\n", "unknown event: ", fmt.Sprintf("%T", e))
	}
}

func (me *UI) WritePrompt() {
	fmt.Fprintf(me.IO, "%s> ", me.CID())
}

func (me *UI) CID() Ident {
	if me.Character == nil {
		return ""
	}
	return me.Character.Ident
}

func (u *UI) ShowIntro() {
	u.Write(u.Center(logo))
	u.Println()
	u.Write(u.Center([]byte("Welcome, to learn more ask for help!")))
	u.Println()
}

func (u *UI) clearScreen() {
	for i := u.rows; i > 0; i-- {
		u.Println()
	}
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
	fmt.Fprintf(me.IO, "\n[%s]: %s\n", id, text)
}

// for notifications
func (me *UI) OtherPlayer(id Ident, text string) {
	fmt.Fprintf(me.IO, "\n%s %s\n", id, text)
}

func (u *UI) showTile(t *Tile) {
	u.CenterPrintln(u.boxed([]byte(t.Short), 40))
	u.Println()
	u.Write(u.Indent(bytes.TrimSpace([]byte(t.Long))))
	u.Println()
	u.Println()
	u.Write(u.Indent([]byte("Exits: ")))
	for d, loc := range t.Nav {
		if loc != "" {
			u.Print(Direction(d).String())
			u.Print(" ")
		}
	}
	u.Println()
	u.Println()
}

func (me *UI) Write(p []byte) (int, error) {
	return me.IO.Write(p)
}

func (u *UI) CenterPrintln(p []byte) {
	u.Write(u.Center(p))
	u.Println()
}

func (me *UI) Println(v ...interface{}) (int, error) {
	p, err := nexus.NewPrinter(me.IO)
	p.Println(v...)
	return int(p.Written), *err
}
func (me *UI) Print(v ...interface{}) (int, error) {
	p, err := nexus.NewPrinter(me.IO)
	p.Print(v...)
	return int(p.Written), *err
}

func (u *UI) boxed(p []byte, width int) []byte {
	var buf bytes.Buffer
	buf.WriteString(frameLine(width))
	scanner := bufio.NewScanner(bytes.NewReader(p))
	for scanner.Scan() {
		line := scanner.Text()
		buf.WriteString("| ")
		buf.WriteString(line)
		suffix := strings.Repeat(" ", width-len(line)-4)
		buf.WriteString(suffix)
		buf.WriteString(" |\n")
	}
	buf.WriteString(frameLine(width))
	return buf.Bytes()
}

func frameLine(width int) string {
	var buf bytes.Buffer
	buf.WriteString("+")
	buf.WriteString(strings.Repeat("-", width-2))
	buf.WriteString("+\n")
	return buf.String()
}

// ----------------------------------------

var nav = map[string]Direction{
	"n":  N,
	"ne": NE,
	"e":  E,
	"se": SE,
	"s":  S,
	"sw": SW,
	"w":  W,
	"nw": NW,
}

var usage = []byte(`
Navigation

n  - north
ne - north-east
e  - east
se - south-east
s  - south
sw - south-west
w  - west
nw - north-west

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
