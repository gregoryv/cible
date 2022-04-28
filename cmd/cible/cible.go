// Command cibtel provides telnet access to a cible game
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/gregoryv/cible"
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
	l := logger.Wrap(log.New(os.Stderr, "", log.LstdFlags))
	if srv {
		g := cible.NewGame()
		g.Logger = l

		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			if err := g.Run(ctx); err != nil {
				g.Log(err)
			}
			cancel() // when game stops, stop the server
		}()

		srv := &cible.Server{Logger: l, Bind: bind}
		if err := srv.Run(ctx, g); err != nil {
			srv.Log(err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// todo use the client
	conn, err := net.Dial("tcp", bind)
	if err != nil {
		l.Log(err)
	}
	// send command

	fmt.Fprintf(conn, os.Getenv("USER")+" join")
}
