// Command cibtel provides telnet access to a cible game
package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"

	. "github.com/gregoryv/cible"
	"github.com/gregoryv/cmdline"
	"github.com/gregoryv/logger"
	"github.com/gregoryv/vt100"
)

var mlog = logger.Wrap(log.Default())

func main() {
	var (
		cli       = cmdline.NewBasicParser()
		bind      = cli.Option("-b, --bind").String(":8089")
		debugFlag = cli.Flag("-d, --debug")
		srv       = cli.Flag("-s, --server")
	)
	cli.Parse()
	if srv {
		defer configureLog(debugFlag)() // configure and defer cleanup

		g := NewGame()
		g.Logger = mlog

		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			if err := g.Run(ctx); err != nil {
				g.Log(err)
			}
			cancel() // when game stops, stop the server
		}()

		srv := &Server{Logger: mlog, Bind: bind}
		if err := srv.Run(ctx, g); err != nil {
			srv.Log(err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	os.Stdout.Write([]byte("\033c"))
	os.Stdout.Write(logo)

	// connect client
	c := NewClient()
	c.Host = bind
	ctx := context.Background()
	if err := c.Connect(ctx); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// create player and join game
	p := Player{Name: Name(os.Getenv("USER"))}
	j := &EventJoin{Player: p}
	c.Out <- NewMessage(j)

	msg := <-c.In
	if err := Decode(j, &msg); err != nil {
		mlog.Log(err)
		os.Exit(1)
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

func configureLog(debugFlag bool) (cleanup func()) {
	w, err := os.OpenFile("server.log", os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	} else {
		log.SetOutput(w)
	}
	if debugFlag {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}
	return func() { _ = w.Close() }
}

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
