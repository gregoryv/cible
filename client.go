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

func Send[T Event](c *Client, e T) error {
	if c.Conn == nil {
		c.Log("send failed: no connection")
		return fmt.Errorf("no connection")
	}

	r := Request{e}
	if err := c.enc.Encode(r); err != nil {
		c.Log(err)
		return err
	}
	c.Logf("send: %T", e)

	if err := c.dec.Decode(&r); err != nil {
		c.Logf("response: %#v", r)
		return err
	}
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
