// Command cibtel provides telnet access to a cible game
package main

import (
	"context"
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
	l := logger.Wrap(log.New(os.Stderr, "", log.LstdFlags|log.Lshortfile))
	if srv {
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

}
