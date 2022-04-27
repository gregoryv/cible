// Command cibtel provides telnet access to a cible game
package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"

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
	l := logger.Wrap(log.New(os.Stderr, "", log.LstdFlags))
	if srv {
		g := NewGame()
		g.Logger = l

		ctx, cancel := context.WithCancel(context.Background())
		go func() {
			if err := g.Run(ctx); err != nil {
				g.Log(err)
			}
			cancel()
		}()

		srv := &Server{
			Logger: l,
			Bind:   bind,
		}
		if err := srv.Run(ctx, g); err != nil {
			srv.Log(err)
			os.Exit(1)
		}
		os.Exit(0)
	}
	conn, err := net.Dial("tcp", bind)
	if err != nil {
		l.Log(err)
	}
	// send command
	// todo maybe use gobs?
	fmt.Fprintf(conn, os.Getenv("USER")+" join")
}

type Server struct {
	logger.Logger
	Bind string
}

func (me *Server) Run(ctx context.Context, g *Game) error {
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
			go me.handleConnection(conn, g)
		}
	}

	return nil
}

func (me *Server) handleConnection(conn net.Conn, g *Game) {
	defer func() {
		_ = recover()
		Trigger(g, Leave("x"))
		conn.Close()
	}()
	me.Log("connect ", conn.RemoteAddr())
	p := make([]byte, 1024)
	for {
		n, err := conn.Read(p)
		if err != nil {
			if err != io.EOF {
				me.Log(err)
			}
			return
		}
		cmd := bytes.TrimRight(p[:n], "\r\n")
		args := strings.Fields(string(cmd))
		switch args[0] {
		case ":q", ":quit":
			return
		case ":join":
			Trigger(g, Join(Player{Name: Name(args[1])}))

		case ":stop game":
			Trigger(g, StopGame())
			return
		}
	}
}
