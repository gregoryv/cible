// Command cibtel provides telnet access to a cible game
package main

import (
	"bytes"
	"fmt"
	"net"
	"os"

	"github.com/gregoryv/cmdline"
	"github.com/gregoryv/logger"
)

func main() {
	var (
		cli  = cmdline.NewBasicParser()
		bind = cli.Option("-b, --bind").String(":8089")
	)
	cli.Parse()
	srv := &TelnetServer{
		Logger: logger.New(),
		Bind:   bind,
	}
	if err := srv.Run(); err != nil {
		srv.Log(err)
		os.Exit(1)
	}
}

type TelnetServer struct {
	logger.Logger
	Bind string
}

func (me *TelnetServer) Run() error {
	ln, err := net.Listen("tcp", me.Bind)
	if err != nil {
		return err
	}
	me.Log("listen on", me.Bind)
	for {
		conn, err := ln.Accept()
		if err != nil {
			me.Log(err)
		}
		go me.handleConnection(conn)
	}
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
		}
	}
}
