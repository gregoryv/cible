package cible

import (
	"encoding/gob"
	"io"
)

type Protocol interface {
	NewEncoder(w io.Writer) Encoder
	NewDecoder(r io.Reader) Decoder
}

// Protocol used over the wire
type GobProtocol struct{}

func (me *GobProtocol) NewEncoder(w io.Writer) Encoder {
	return gob.NewEncoder(w)
}

func (me *GobProtocol) NewDecoder(r io.Reader) Decoder {
	return gob.NewDecoder(r)
}

type Encoder interface {
	Encode(v any) error
}

type Decoder interface {
	Decode(v any) error
}
