package cible

import (
	"context"
	"encoding/gob"
	"fmt"
	"io"
	"net"

	"github.com/gregoryv/logger"
)

func NewClient() *Client {
	return &Client{
		Logger: logger.Silent,
	}
}

type Client struct {
	logger.Logger
	Host string

	net.Conn

	enc *gob.Encoder
	dec *gob.Decoder
}

func (me *Client) Connect(ctx context.Context) error {
	conn, err := net.Dial("tcp", me.Host)
	if err != nil {
		return err
	}
	me.Conn = conn
	me.Log("connected to", me.Host)
	me.enc = gob.NewEncoder(conn)
	me.dec = gob.NewDecoder(conn)

	return nil
}

func Send[T Event](c *Client, e T) {
	if c.Conn == nil {
		e.setErr(fmt.Errorf("no connection"))
		c.Log("send failed: no connection")
		return
	}

	r := Request{e}
	if err := c.enc.Encode(r); err != nil {
		c.Log(err)
	}
	// todo wait for response and update event
}

// ----------------------------------------

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
			Trigger(g, StopGame()).Done()
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
		_ = recover()
		Trigger(g, Leave("x"))
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

		switch e := r.Event.(type) {
		case EventJoin:
			// the incomming event is inactive, create a new
			j := Trigger(g, e.New())
			if err := j.Done(); err != nil {
				me.Log(err)
				continue
			}
			// todo why doesn't the trigger affect the game??
			// todo respond

		default:
			me.Log("unknown event")
			continue
		}
		// todo send response
		_ = enc

	}
}

type Request struct {
	Event interface{} // if only Event then we cannot gob encode it
}

func init() {
	gob.Register(EventJoin{})
}
