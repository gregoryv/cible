package cible

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"net"

	"github.com/gregoryv/logger"
	"github.com/gregoryv/nexus"
)

func NewClient() *Client {
	return &Client{
		Logger: logger.Silent,
	}
}

type Client struct {
	logger.Logger
	Host string

	Up   net.Conn // client to server
	Down net.Conn // server to client

	enc *gob.Encoder
	dec *gob.Decoder
}

func (me *Client) Connect(ctx context.Context) error {
	up, err := net.Dial("tcp", me.Host)
	if err != nil {
		return err
	}
	me.Up = up
	me.Log("connected to", me.Host)
	me.enc = gob.NewEncoder(up)
	me.dec = gob.NewDecoder(up)

	down, err := net.Dial("tcp", me.Host)
	me.Down = down
	return err
}

func (me *Client) Close() {
	me.Up.Close()
	me.Down.Close()
}

func (c *Client) CheckState() error {
	if c.Up == nil {
		return fmt.Errorf("client is disconnected")
	}
	return nil
}

// Sends an event and waits for the response
func Send[T any](c *Client, e *T) (x T, err error) {
	next := nexus.NewStepper(&err)
	next.Step(func() { err = c.CheckState() })

	msg := NewMessage(e)
	next.Stepf("write on wire: %w", func() {
		err = c.enc.Encode(&msg)
	})

	next.Stepf("read response: %w", func() {
		err = c.dec.Decode(&msg)
	})

	next.Stepf("response: %w", func() {
		err = msg.CheckError()
	})

	next.Stepf("decode body: %w", func() {
		r := bytes.NewReader(msg.Body)
		err = gob.NewDecoder(r).Decode(&x)
	})
	return
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

// Message is for transferring events between client and server using
// encoding/gob
type Message struct {
	EventName string
	Body      []byte
}

func (m *Message) String() string {
	return fmt.Sprintf(
		"message %s %v bytes", m.EventName, m.Size(),
	)
}

func (m *Message) Size() int {
	var buf bytes.Buffer
	gob.NewEncoder(&buf).Encode(*m)
	return buf.Len()
}

func (m *Message) CheckError() error {
	if m.EventName == "error" {
		return fmt.Errorf("%s", string(m.Body))
	}
	return nil
}
