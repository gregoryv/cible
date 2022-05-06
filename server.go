package cible

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/gregoryv/logger"
)

func NewServer() *Server {
	return &Server{
		Logger:          logger.Silent,
		Bind:            "",
		MaxConnections:  100,
		MaxAcceptErrors: 100,
	}
}

type Server struct {
	logger.Logger
	Bind            string
	MaxConnections  int // not really max allowed players, more like DOS throttling
	MaxAcceptErrors int

	net.Listener

	game *Game
}

func (me *Server) Run(ctx context.Context, g *Game) error {
	if me.Listener == nil {
		ln, err := net.Listen("tcp", me.Bind)
		if err != nil {
			return err
		}
		me.Listener = ln
		me.Log("server listen on", ln.Addr())
	}
	c := make(chan net.Conn, me.MaxConnections)
	acceptErr := make(chan error, 1)
	max := me.MaxAcceptErrors

	go func() {
		backoff := 20 * time.Millisecond
		for {
			conn, err := me.Listener.Accept()
			if err != nil {
				me.Log(err)
				max--
				if max < 0 {
					me.Log("to many accept errors")
					acceptErr <- err // signal connect loop we are done
				}
				backoff *= 2
				<-time.After(backoff)
				continue
			}
			backoff = time.Millisecond
			c <- conn
		}
	}()

	me.game = g

	gobish := &GobProtocol{}
connectLoop:
	for {
		select {
		case <-ctx.Done():
			me.Log("server interrupted")
			break connectLoop
		case conn := <-c:
			go func() {
				me.Log("connect ", conn.RemoteAddr())
				tr := NewTransceiver(conn, gobish)
				if err := me.communicate(tr); err != nil {
					me.Log(err)
				}
				conn.Close()
			}()
		case err := <-acceptErr:
			me.Log("server stopped")
			return fmt.Errorf("exceeded max accept errors %v: %w", me.MaxAcceptErrors, err)
		}
	}
	return nil
}

func (me *Server) communicate(tr *Transceiver) error {
	var cid Ident // set on first EventJoin
	defer func() {
		// graceful panic handling
		if e := recover(); e != nil {
			me.Log(e)
		}
		me.game.Do(Leave(cid))
		me.Log(cid, " disconnected")
	}()

	for {
		var msg Message
		if err := tr.Receive(&msg); err != nil {
			if err != io.EOF {
				return err
			}
			return nil
		}
		me.Logf("recv %s", msg.String())

		x, known := NewNamedEvent(msg.EventName)
		if !known {
			continue
		}

		dec := gob.NewDecoder(bytes.NewReader(msg.Body))
		if err := dec.Decode(x); err != nil {
			me.Log(err)
		}

		// new player joined, set the transceiver for further
		// communication
		if e, ok := x.(*EventJoin); ok {
			e.tr = tr
		}

		if e, ok := x.(Event); ok {
			if err := me.game.Do(e); err != nil {
				msg.Body = []byte(err.Error())
			}
		}
		// ignore other events
	}
}

type needsTransmitter interface {
	setTransmitter(Transmitter)
}
