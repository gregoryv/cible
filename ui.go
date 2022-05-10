package cible

import (
	"bufio"
	"bytes"
	"context"
	_ "embed"
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

	Character
	Location string // used in prompt

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
	send <- NewMessage(&EventJoinGame{
		Player: player,
	})

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
loop:
	for {
		// notify the prompt to update when the events have stopped
		promptUpdate <- struct{}{}

		// use a selct so that we only process one event at the time,
		// either incoming or outgoing
		select {
		case <-ctx.Done():
			return nil

		case m := <-u.in:
			// handle incoming messages
			e, known := NewEvent(m.EventName)
			if !known {
				continue
			}
			Decode(e, &m)
			u.HandleEvent(e)

		case input := <-u.playerInput:
			switch input {
			case "": // ignore empty
			case "n", "ne", "e", "se", "s", "sw", "w", "nw":
				send <- NewMessage(&EventMove{Direction: nav[input]})

			case "l", "look":
				send <- NewMessage(&EventLook{})

			case "i", "inventory":
				u.showInventory()

			case "h", "help":
				u.showUsage()

			case "q", "quit":
				send <- NewMessage(&EventLeave{})
				<-time.After(40 * time.Millisecond)
				p.Println("\nBye!")
				return nil

			default:
				fields := strings.Fields(input)
				switch fields[0] {
				case "p", "pickup":
					if len(fields) == 1 {
						u.Println("pickup what?")
						continue loop
					}
					send <- NewMessage(&EventPickup{
						Item: Item{
							Name: Name(fields[1]),
						},
					})

				default:
					if input != "" {
						send <- NewMessage(&EventSay{Text: input})
					}
				}
			}
		}
	}
}

func (u *UI) HandleEvent(e interface{}) {
	switch e := e.(type) {
	case *EventInventoryUpdate:
		u.Character.Inventory = *e.Inventory

	case *EventGoAway:
		u.OtherPlayer(e.Name, "went away")

	case *EventApproach:
		u.OtherPlayer(e.Name, "is near")

	case *EventSay:
		u.OtherPlayerSays(e.Name, e.Text)

	case *EventJoin:
		u.OtherPlayer(e.Name, "joined game")

	case *EventJoinGame:
		// when you coin
		u.Character = *e.Character
		u.Location = fmt.Sprintf("%s/%s", e.Title, e.Position.Tile)
		u.Write(Center(
			[]byte(
				"You have entered the " + e.Title + " area",
			),
		))
		u.Println()
		u.Println()
		u.Println()

	case *EventLeave:
		u.OtherPlayer(e.Name, "left game")

	case *EventLook:
		u.showTile(&e.Tile, true)
		u.Println()
		if len(e.Loose) > 0 {
			u.Println()
		}
		for _, item := range e.Loose {
			u.Write(Center([]byte("You found a " + item.Name + "!")))
		}
		u.showNav(&e.Tile.Nav)
		u.Println()

	case *EventMove:
		if u.Character.Position.Equal(e.Position) {
			u.Println("cannot move in that direction")
			return
		}
		u.Character.Position = e.Position
		u.showTile(e.Tile, false)
		u.Println()
		u.Location = fmt.Sprintf("%s/%s", e.Title, e.Position.Tile)

	default:
		u.Println("\n", "unknown event: ", fmt.Sprintf("%T", e))
	}
}

func (u *UI) WritePrompt() {
	fmt.Fprintf(u.IO, "%s@%s> ", u.Character.Name, strings.ToLower(u.Location))
}

func (u *UI) showInventory() {
	u.Println()
	u.Write(Center(Boxed([]byte("Inventory"), 40)))
	u.Println()
	var buf bytes.Buffer
	for i, item := range u.Character.Inventory.Items {
		switch {
		case item.Count > 1:
			buf.WriteString(fmt.Sprintf("%v. %v %-30s\n", i+1, item.Count, item.Name))
		case item.Count == 1:
			buf.WriteString(fmt.Sprintf("%v. %-30s\n", i+1, item.Name))
		}
	}
	u.Write(Indent(buf.Bytes()))
	u.Println()
}

func (u *UI) showUsage() {
	u.Write(Center(Boxed(usage, 55)))
	u.Println()
}

func (u *UI) ShowIntro() {
	u.Write(Center(logo))
	u.Println()
	u.Println()
	u.Write(Center("To learn more, just ask for help!"))
	u.Println(strings.Repeat("\n", 8))
}

func (u *UI) showTile(t *Tile, long bool) {
	u.Println()
	u.Write(Center(Boxed(CenterIn([]byte(t.Short), 36), 40)))
	if long {
		u.Println()
		u.Println()
		u.Write(Indent(
			bytes.TrimSpace([]byte(t.Long)),
		))
	}
}

func (u *UI) showNav(n *Nav) {
	u.Println()
	u.Println()
	u.Write(Indent(exits(*n)))
	u.Println()
	u.Println()
}

func exits(n Nav) []byte {
	var buf bytes.Buffer
	for d, loc := range n {
		if loc != "" {
			buf.WriteString(Direction(d).String())
			buf.WriteString(" ")
		}
	}
	return bytes.TrimRight(buf.Bytes(), " ")
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

func (me *UI) Write(p []byte) (int, error) {
	return me.IO.Write(p)
}

func (me *UI) Println(v ...interface{}) (int, error) {
	p, err := nexus.NewPrinter(me.IO)
	p.Println(v...)
	return int(p.Written), *err
}

func (me *UI) Printf(format string, v ...interface{}) (int, error) {
	p, err := nexus.NewPrinter(me.IO)
	p.Printf(format, v...)
	return int(p.Written), *err
}

func (me *UI) Print(v ...interface{}) (int, error) {
	p, err := nexus.NewPrinter(me.IO)
	p.Print(v...)
	return int(p.Written), *err
}

func Boxed(p []byte, minWidth int) []byte {
	scanner := bufio.NewScanner(bytes.NewReader(p))
	var buf bytes.Buffer
	buf.WriteString(frameLine(minWidth))
	for scanner.Scan() {
		line := scanner.Text()
		buf.WriteString("| ")
		buf.WriteString(line)
		r := minWidth - len(line) - 4
		var suffix string
		if r >= 0 {
			suffix = strings.Repeat(" ", r)
		}
		buf.WriteString(suffix)
		buf.WriteString(" |\n")
	}
	buf.WriteString(frameLine(minWidth))
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

//go:embed asset/usage.txt
var usage []byte

//go:embed asset/logo.txt
var logo []byte
