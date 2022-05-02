package main

import (
	"bufio"
	"context"
	"fmt"
	"os"

	. "github.com/gregoryv/cible"
	"github.com/gregoryv/vt100"
)

func NewUI() *UI {
	return &UI{}
}

type UI struct {
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
				mlog.Log(err)
				os.Exit(1)
			}
			playerInput <- scanner.Text()
		}
	}()

	var (
		fg     = vt100.ForegroundColors()
		yellow = fg.Yellow.Bytes()
		cyan   = fg.Cyan.Bytes()

		vt    = vt100.Attributes()
		reset = vt.Reset.Bytes()
	)

	// handle incoming messages
	for {
		select {
		case m := <-c.In:
			e, known := NewNamedEvent(m.EventName)
			if !known {
				continue
			}
			Decode(e, &m)
			switch e := e.(type) {
			case *EventSay:
				fmt.Printf("\n%s%s: %s%s\n", cyan, e.Ident, e.Text, reset)
			default:
				fmt.Printf("\n%s%v%s\n", yellow, e, reset)
			}
			writePrompt()

		case input := <-playerInput:
			switch input {
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
				c.Close()
				fmt.Println("bye")
				os.Exit(0)
			default:
				e := &EventSay{Ident: cid, Text: input}
				c.Out <- NewMessage(e)
			}

		}
	}
}
