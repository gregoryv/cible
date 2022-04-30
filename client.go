package cible

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"net"

	"github.com/google/uuid"
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

func (c *Client) CheckState() error {
	if c.Conn == nil {
		return fmt.Errorf("client is disconnected")
	}
	return nil
}

// Sends an event and waits for the response
func Send[T any](c *Client, e *T) (x T, err error) {

	// Make it easier to verify the flow with step functions, as any
	// error results in a return.
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
		Id:        uuid.NewString(),
		EventName: fmt.Sprintf("%T", *v),
		Body:      buf.Bytes(),
	}
}

// Message is for transferring events between client and server using
// encoding/gob
type Message struct {
	Id        string
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
