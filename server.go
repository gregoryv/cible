package cible

import (
	"context"
	"encoding/gob"
	"io"
	"net"

	"github.com/gregoryv/logger"
)

func NewServer() *Server {
	return &Server{
		Logger:         logger.Silent,
		Bind:           "",
		MaxConnections: 100,
	}
}

type Server struct {
	logger.Logger
	Bind           string
	MaxConnections int

	net.Addr // set after running

	game *Game
}

func (me *Server) Run(ctx context.Context, g *Game) error {
	ln, err := net.Listen("tcp", me.Bind)
	if err != nil {
		return err
	}
	me.Addr = ln.Addr()
	me.Log("server listen on", ln.Addr())

	c := make(chan net.Conn, me.MaxConnections)
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				me.Log(err)
				continue // todo this may spin out of control
			}
			c <- conn
		}
	}()

	me.game = g
connectLoop:
	for {
		select {
		case <-ctx.Done():
			break connectLoop
		case conn := <-c:
			go me.handleConnection(conn, g)
		}
	}

	me.Log("server closed")
	return nil
}

func (me *Server) handleConnection(conn net.Conn, g *Game) {
	defer func() {
		// graceful connection handling
		me.Log(recover())
		conn.Close()
	}()
	me.Log("connect ", conn.RemoteAddr())

	enc := gob.NewEncoder(conn)
	dec := gob.NewDecoder(conn)

	for {
		var r Request
		if err := dec.Decode(&r); err != nil {
			if err != io.EOF {
				me.Log(err)
			}
			return
		}
		me.Logf("recv: %T", r.Event)
		// todo figure out how to call Trigger on everything that comes in

		switch e := r.Event.(type) {
		case EventJoin:
			task, x := Trigger(g, &e)
			if err := task.Done(); err != nil {
				me.Log(err)
				continue
			}
			r.Event = x
		}

		me.Logf("send: %#v", r.Event)
		if err := enc.Encode(r); err != nil {
			me.Log(err)
		}

	}
}
