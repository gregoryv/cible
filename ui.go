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
	DefaultTextFormat.cols = cols

	return &UI{
		Logger: logger.Silent,

		IO:          NewRWCache(NewStdIO()),
		playerInput: make(chan string, 1),

		out: make(chan Message, 1),
		in:  make(chan Message, 1),

		cols: cols,
		rows: rows,
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
	send <- NewMessage(&PlayerJoin{Player: player})

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
				u.showUsage()

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
		u.OtherPlayerSays(e.Name, e.Text)

	case *EventJoin:
		u.OtherPlayer(e.Name, "joined")

	case *PlayerJoin:
		u.Character = e.Character

	case *EventLeave:
		u.OtherPlayer(e.Name, "left game")

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
	fmt.Fprintf(me.IO, "%s> ", me.CharacterName())
}

func (me *UI) CID() Ident {
	if me.Character == nil {
		return ""
	}
	return me.Character.Ident
}

func (me *UI) CharacterName() Name {
	if me.Character == nil {
		return ""
	}
	return me.Character.Name
}

func (u *UI) showUsage() {
	u.Write(Center(Boxed(usage, 40)))
	u.Println()
}

func (u *UI) ShowIntro() {
	u.Write(Center(logo))
	u.Println()
	u.Write(Center("Welcome, to learn more ask for help!"))
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
func (me *UI) OtherPlayerSays(name Name, text string) {
	fmt.Fprintf(me.IO, "\n %s: %s\n", name, text)
}

// for notifications
func (me *UI) OtherPlayer(name Name, text string) {
	fmt.Fprintf(me.IO, "\n%s %s\n", name, text)
}

func (u *UI) showTile(t *Tile) {
	u.Write(Center(Boxed(CenterIn([]byte(t.Short), 36), 40)))
	u.Println()
	u.Println()
	u.Write(Indent(
		bytes.TrimSpace([]byte(t.Long)),
	))
	u.Println()
	u.Println()
	u.Write(Indent(exits(t.Nav)))
	u.Println()
	u.Println()
}

func exits(n Nav) []byte {
	var buf bytes.Buffer
	buf.WriteString("Exits: ")
	for d, loc := range n {
		if loc != "" {
			buf.WriteString(Direction(d).String())
			buf.WriteString(" ")
		}
	}
	return bytes.TrimRight(buf.Bytes(), " ")
}

func (me *UI) Write(p []byte) (int, error) {
	return me.IO.Write(p)
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

func Boxed(p []byte, width int) []byte {
	scanner := bufio.NewScanner(bytes.NewReader(p))
	var buf bytes.Buffer
	buf.WriteString(frameLine(width))
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
