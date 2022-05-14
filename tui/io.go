package tui

import (
	"bytes"
	"io"
	"os"
)

type IO io.ReadWriter

func NewRWCache(rw io.ReadWriter) *RWCache {
	return &RWCache{
		ReadWriter: rw,
	}
}

type RWCache struct {
	io.ReadWriter
	LastRead  []byte
	LastWrite []byte
}

func (me *RWCache) Read(p []byte) (int, error) {
	n, err := me.ReadWriter.Read(p)
	me.LastRead = p
	return n, err
}

func (me *RWCache) Write(p []byte) (int, error) {
	n, err := me.ReadWriter.Write(p)
	me.LastWrite = p
	return n, err
}

// ----------------------------------------

func NewBufIO() *BufIO {
	return &BufIO{
		input:  &bytes.Buffer{},
		output: &bytes.Buffer{},
	}
}

type BufIO struct {
	input  *bytes.Buffer
	output *bytes.Buffer
}

func (me *BufIO) Read(p []byte) (int, error) {
	return me.input.Read(p)
}

func (me *BufIO) Write(p []byte) (int, error) {
	return me.output.Write(p)
}

func NewStdIO() *StdIO {
	return &StdIO{
		Reader: os.Stdin,
		Writer: os.Stdout,
	}
}

type StdIO struct {
	io.Reader // input
	io.Writer // output
}
