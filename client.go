package cible

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/gregoryv/logger"
)

func NewClient() *Client {
	return &Client{
		Logger: logger.Silent,
		Out:    make(chan Message, 1),
		In:     make(chan Message, 1),
	}
}

type Client struct {
	logger.Logger
	Host string

	net.Conn

	Out chan Message
	In  chan Message
}

func (me *Client) Connect(ctx context.Context) error {
	conn, err := net.Dial("tcp", me.Host)
	if err != nil {
		return err
	}
	me.Conn = conn
	me.Log("connected to", me.Host)
	tr := NewTransceiver(conn, &GobProtocol{})

	// transmit outgoing messages
	go func() {
		for {
			select {
			case <-ctx.Done():
			case m := <-me.Out:
				if err := tr.Transmit(m); err != nil {
					me.Log(err)
					return
				}
			}
		}
	}()

	// receive incoming messages
	go func() {
		for {
			var msg Message
			if err := tr.Receive(&msg); err != nil {
				me.Log(err)
				return
			}
			me.Log("in:", msg.String())
			me.In <- msg
		}
	}()

	<-time.After(20 * time.Millisecond)
	return nil
}

func (c *Client) CheckState() error {
	if c.Conn == nil {
		return fmt.Errorf("client disconnected")
	}
	return nil
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
	id, _ := uuid.Parse(m.Id)
	return fmt.Sprintf(
		"%s[%v] %v bytes", m.EventName, id.ID(), m.Size(),
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

func Decode(v interface{}, m *Message) error {
	dec := gob.NewDecoder(bytes.NewReader(m.Body))
	if err := dec.Decode(v); err != nil {
		if err != io.EOF {
			return err
		}
	}
	return nil
}
