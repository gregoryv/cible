package cible

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/gregoryv/logger"
)

func NewClient() *Client {
	return &Client{}
}

type Client struct {
	logger.Logger
	Host string
}

func (me *Client) Connect(ctx context.Context) error {
	conn, err := net.Dial("tcp", me.Host)
	if err != nil {
		return err
	}

	// send command
	// todo maybe use gobs?
	fmt.Fprintf(conn, os.Getenv("USER")+" join")

	for {
		select {
		case <-ctx.Done():
		}
	}

	return conn.Close()
}

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
				continue
			}
			c <- conn
		}
	}()

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
