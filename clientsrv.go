package cible

import (
	"bytes"
	"context"
	"io"
	"net"
	"strings"

	"github.com/gregoryv/logger"
)

func NewServer() *Server {
	return &Server{
		Logger: logger.Silent,
		Bind:   "",
	}
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
	me.Log("server listen on", ln.Addr())

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
