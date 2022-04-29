package cible

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
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
	var cid Ident // set on first EventJoin
	defer func() {
		// graceful connection handling
		e := recover()
		if e != nil {
			me.Log(e)
		}
		Trigger(g, Leave(cid))
		me.Log(cid, " disconnected")
		conn.Close()
	}()
	me.Log("connect ", conn.RemoteAddr())

	enc := gob.NewEncoder(conn)
	dec := gob.NewDecoder(conn)

	for {
		var msg Message
		if err := dec.Decode(&msg); err != nil {
			if err != io.EOF {
				me.Log(err)
			}
			return
		}
		me.Logf("recv %s", msg.String())
		x, found := newNamedEvent(msg.EventName)
		if !found {
			err := fmt.Errorf("missing named event %s", msg.EventName)
			msg.Body = []byte(err.Error())
			me.Log(err)
			msg.EventName = "error"
		} else {

			dec := gob.NewDecoder(bytes.NewReader(msg.Body))
			if err := dec.Decode(x); err != nil {
				me.Log(err)
			}

			task, x := Trigger(g, (x).(Event))
			if err := task.Done(); err != nil {
				me.Log(err)
				continue
			}
			if msg.EventName == "cible.EventJoin" {
				cid = x.(*EventJoin).Ident
			}
			var buf bytes.Buffer
			if err := gob.NewEncoder(&buf).Encode(x); err != nil {
				me.Log(err)
			}
			msg.Body = buf.Bytes()
		}
		me.Logf("send %v", msg.String())
		if err := enc.Encode(msg); err != nil {
			me.Log(err)
		}
	}
}
