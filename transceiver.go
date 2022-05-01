package cible

import (
	"fmt"
	"io"
)

func NewTransceiver(rw io.ReadWriter, proto Protocol) *Transceiver {
	return &Transceiver{
		proto.NewEncoder(rw),
		proto.NewDecoder(rw),
	}
}

type Transceiver struct {
	Encoder
	Decoder
}

func (me *Transceiver) Transmit(v any) error {
	if me == nil {
		return fmt.Errorf("transceiver not set")
	}
	return me.Encode(v)
}

func (me *Transceiver) Receive(v any) error {
	if me == nil {
		return fmt.Errorf("transceiver not set")
	}
	return me.Decode(v)
}

type Transmitter interface {
	Transmit(any) error
}
