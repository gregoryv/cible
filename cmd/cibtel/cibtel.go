// Command cibtel provides telnet access to a cible game
package main

import (
	"bytes"
	"context"
	"fmt"
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
	)
	cli.Parse()

	g := cible.NewGame()
	g.Logger = logger.New()

	srv := &TelnetServer{
		Logger: logger.New(),
		Bind:   bind,
		Game:   g,
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		if err := g.Run(ctx); err != nil {
			g.Log(err)
		}
		cancel()
	}()

	if err := srv.Run(ctx); err != nil {
		srv.Log(err)
		os.Exit(1)
	}
}

type TelnetServer struct {
	logger.Logger
	Bind string
	*cible.Game
}

func (me *TelnetServer) Run(ctx context.Context) error {
	ln, err := net.Listen("tcp", me.Bind)
	if err != nil {
		return err
	}
	me.Log("listen on", me.Bind)

	c := make(chan net.Conn)
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				me.Log(err)
				continue
			}
			c <- conn
		}
	}()

	for {
		select {
		case <-ctx.Done():
			me.Log("server stop")
			return nil
		case conn := <-c:
			go me.handleConnection(conn)
		}
	}

	return nil
}

func (me *TelnetServer) handleConnection(conn net.Conn) {
	me.Log("connect", conn.RemoteAddr())
	p := make([]byte, 1024)
	for {
		n, err := conn.Read(p)
		if err != nil {
			me.Log(err)
		}
		cmd := bytes.TrimRight(p[:n], "\r\n")
		fmt.Println(string(cmd))
		switch string(cmd) {
		case ":q", ":quit":
			conn.Close()
			return
		case ":stop game":
			cible.Trigger(me.Game, cible.StopGame())
			return
		}
	}
}
