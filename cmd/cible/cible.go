// Command cibtel provides telnet access to a cible game
package main

import (
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
	c := NewClient()
	c.Host = bind
	ctx := context.Background()
	ui := NewUI()
	if err := ui.Run(ctx, c); err != nil {
		fmt.Println(err)
		os.Exit(1)
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
