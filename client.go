package cible

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
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

func Send[T any](c *Client, e *T) (T, error) {
	if c.Conn == nil {
		c.Log("send failed: no connection")
		return *e, fmt.Errorf("no connection")
	}

	var buf bytes.Buffer
	gob.NewEncoder(&buf).Encode(*e)
	r := Request{
		EventName: fmt.Sprintf("%T", *e),
		Body:      buf.Bytes(),
	}
	if err := c.enc.Encode(&r); err != nil {
		c.Log(err)
		return *e, err
	}
	c.Logf("send %s, body %v bytes", r.EventName, len(r.Body))

	if err := c.dec.Decode(&r); err != nil {
		c.Log(err)
		return *e, err
	}
	var x T
	dec := gob.NewDecoder(bytes.NewReader(r.Body))
	if err := dec.Decode(&x); err != nil {
		c.Log(err)
		return *e, err
	}
	c.Logf("recv %s, body %v bytes", r.EventName, len(r.Body))
	return x, nil
}

// ----------------------------------------

type Request struct {
	EventName string
	Body      []byte
}

func init() {
	// todo all events must be registerd for transfer via request
	gob.Register(EventJoin{})
	gob.Register(Request{})
}
