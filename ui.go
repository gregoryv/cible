package cible

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/gregoryv/logger"
	"github.com/gregoryv/vt100"
)

func NewUI() *UI {
	return &UI{
		Logger: logger.Silent,
	}
}

type UI struct {
	logger.Logger
	*Client
}

func (me *UI) Run(ctx context.Context, c *Client) error {
	os.Stdout.Write([]byte("\033c"))
	os.Stdout.Write(logo)

	// connect client
	if err := c.Connect(ctx); err != nil {
		return err
	}

	// create player and join game
	p := Player{Name: Name(os.Getenv("USER"))}
	j := &EventJoin{Player: p}
	c.Out <- NewMessage(j)

	msg := <-c.In
	if err := Decode(j, &msg); err != nil {
		return err
	}

	cid := j.Ident
	// uggly way to set current pos, todo fix it
	m := MoveCharacter(cid, N)
	c.Out <- NewMessage(m)
	m.Direction = S
	c.Out <- NewMessage(m)

	writePrompt := func() { fmt.Printf("%s> ", cid) }
	playerInput := make(chan string, 1)
	go func() {
		for {
			writePrompt()
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
			if err := scanner.Err(); err != nil {
				me.Log(err)
				os.Exit(1)
			}
			playerInput <- scanner.Text()
		}
	}()

	// handle incoming messages
	for {
		select {
		case m := <-c.In:
			e, known := NewNamedEvent(m.EventName)
			if !known {
				continue
			}
			Decode(e, &m)
			if e, ok := e.(interface{ AffectUI(*UI) }); ok {
				e.AffectUI(me)
			} else {
				fmt.Printf("\n%s%v%s\n", yellow, e, reset)
			}
			writePrompt()

		case input := <-playerInput:
			switch input {
			case "": // ignore
			case "n", "w", "s", "e":
				mv := MoveCharacter(cid, nav[input])
				c.Out <- NewMessage(mv)

			case "l":
				// todo first position
				if m.Tile != nil {
					os.Stdout.Write([]byte(m.Tile.Long))
					fmt.Println()
				}
			case "h", "help":
				os.Stdout.Write(usage)
			case "q":
				c.Out <- NewMessage(Leave(cid))
				<-time.After(40 * time.Millisecond)
				c.Close()
				fmt.Println("\nBye!")
				os.Exit(0)
			default:
				e := &EventSay{Ident: cid, Text: input}
				c.Out <- NewMessage(e)
			}

		}
	}
}

// only for speach
func (me *UI) OtherPlayerSays(id Ident, text string) {
	fmt.Printf("\n%s%s: %s%s\n", cyan, id, text, reset)
}

// for notifications
func (me *UI) OtherPlayer(id Ident, text string) {
	fmt.Printf("\n%s%s: %s%s\n", yellow, id, text, reset)
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
