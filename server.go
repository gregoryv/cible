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

connectLoop:
	for {
		select {
		case <-ctx.Done():
			me.Log("server interrupted")
			break connectLoop
		case conn := <-c:
			go func() {
				me.Log("connect ", conn.RemoteAddr())
				me.communicate(conn)
				conn.Close()
			}()
		case err := <-acceptErr:
			me.Log("server stopped")
			return fmt.Errorf("exceeded max accept errors %v: %w", me.MaxAcceptErrors, err)
		}
	}
	return nil
}

func (me *Server) communicate(conn io.ReadWriter) {
	var cid Ident // set on first EventJoin
	defer func() {
		// graceful panic handling
		if e := recover(); e != nil {
			me.Log(e)
		}
		me.game.Do(Leave(cid))
		me.Log(cid, " disconnected")
	}()

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

			if err := me.game.Do((x).(Event)); err != nil {
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
