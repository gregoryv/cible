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

func main() {
	var (
		cli  = cmdline.NewBasicParser()
		bind = cli.Option("-b, --bind").String(":8089")
		srv  = cli.Flag("-s, --server")
	)
	cli.Parse()
	l := logger.Silent
	if srv {
		l = logger.Wrap(log.New(os.Stderr, "", log.LstdFlags|log.Lshortfile))
		g := NewGame()
		g.Logger = l

		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			if err := g.Run(ctx); err != nil {
				g.Log(err)
			}
			cancel() // when game stops, stop the server
		}()

		srv := &Server{Logger: l, Bind: bind}
		if err := srv.Run(ctx, g); err != nil {
			srv.Log(err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	os.Stdout.Write(logo)
	// connect client
	c := NewClient()
	c.Logger = l
	c.Host = bind
	ctx := context.Background()
	if err := c.Connect(ctx); err != nil {
		l.Log(err)
		os.Exit(1)
	}

	// send command
	p := Player{Name: Name(os.Getenv("USER"))}
	j, err := Send(c, &EventJoin{Player: p})
	if j.Ident == "" {
		l.Log("join failed, missing ident", err)
		os.Exit(1)
	}
	cid := j.Ident

	var m Movement
	for {
		fmt.Printf("%s> ", p.Name)
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		err := scanner.Err()
		if err != nil {
			l.Log(err)
			os.Exit(1)
		}
		// todo handle commands
		input := scanner.Text()

		switch input {
		case "n", "w", "s", "e":
			m, err = Send(c, MoveCharacter(cid, nav[input]))
			fmt.Println(m.Direction, " => ", m.Tile.Short)

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
