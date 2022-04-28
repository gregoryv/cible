package cible

import (
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

func Send(c *Client, r *Request) error {
	if c.Conn == nil {
		c.Log("send failed: no connection")
		return fmt.Errorf("no connection")
	}

	if err := c.enc.Encode(r); err != nil {
		c.Log(err)
		return err
	}
	c.Logf("send: %T", r.Event)

	if err := c.dec.Decode(&r); err != nil {
		c.Log(err)
		return err
	}
	c.Logf("received: %#v", r.Event)
	return nil
	// todo wait for response and update event
}

// ----------------------------------------

type Request struct {
	Event interface{} // if only Event then we cannot gob encode it
}

func init() {
	gob.Register(EventJoin{})
}
