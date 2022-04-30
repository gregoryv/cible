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
	j, err := Send(c, &EventJoin{Player: p})
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	cid := j.Ident

	var m Movement
	// uggly way to set current pos, todo fix it
	m, _ = Send(c, MoveCharacter(cid, N))
	m, _ = Send(c, MoveCharacter(cid, S))

	for {
		fmt.Printf("%s> ", p.Name)
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		err := scanner.Err()
		if err != nil {
			mlog.Log(err)
			os.Exit(1)
		}

		input := scanner.Text()
		switch input {
		case "n", "w", "s", "e":
			m, err = Send(c, MoveCharacter(cid, nav[input]))
			if err == nil {
				fmt.Println(m.Direction, " => ", m.Tile.Short)
			}

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
			err = fmt.Errorf("?")
		}
		if err != nil {
			fmt.Println(err)
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
