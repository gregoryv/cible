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
		return *e, fmt.Errorf("no connection")
	}

	// write message on the wire
	msg := NewMessage(e)
	if err := c.enc.Encode(&msg); err != nil {
		return *e, err
	}

	// read response, reuse same message
	if err := c.dec.Decode(&msg); err != nil {
		return *e, err
	}
	if msg.EventName == "error" {
		err := fmt.Errorf("%s", string(msg.Body))
		return *e, err
	}

	// decode the body and return the same type
	var x T
	dec := gob.NewDecoder(bytes.NewReader(msg.Body))
	if err := dec.Decode(&x); err != nil {
		return *e, err
	}
	return x, nil
}

// ----------------------------------------

func NewMessage[T any](v *T) Message {
	var buf bytes.Buffer
	gob.NewEncoder(&buf).Encode(*v)
	return Message{
		EventName: fmt.Sprintf("%T", *v),
		Body:      buf.Bytes(),
	}
}

type Message struct {
	EventName string
	Body      []byte
}

func (m *Message) String() string {
	return fmt.Sprintf(
		"message %s, body %v bytes", m.EventName, len(m.Body),
	)
}
