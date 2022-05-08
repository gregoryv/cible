package cible

import (
	"bufio"
	"bytes"
	"strings"
)

type TextFormat struct {
	cols int
}

func (f *TextFormat) Indent(p []byte) []byte {
	s := f.cols / 8
	indent := strings.Repeat(" ", s)

	scanner := bufio.NewScanner(bytes.NewReader(p))
	var buf bytes.Buffer
	for scanner.Scan() {
		line := scanner.Text()
		buf.WriteString(indent)
		buf.WriteString(line)
		buf.WriteString("\n")
	}
	return bytes.TrimRight(buf.Bytes(), "\n")
}

func (f *TextFormat) Center(p []byte) []byte {
	scanner := bufio.NewScanner(bytes.NewReader(p))
	var buf bytes.Buffer
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < f.cols {
			prefix := strings.Repeat(" ", (f.cols-len(line))/2)
			buf.WriteString(prefix)
			buf.WriteString(line)
		}
		buf.WriteString("\n")
	}
	return bytes.TrimRight(buf.Bytes(), "\n")
}

var DefaultTextFormat = &TextFormat{
	cols: 72,
}

func Center(p interface{}) []byte {
	switch p := p.(type) {
	case []byte:
		return DefaultTextFormat.Center(p)
	case string:
		return DefaultTextFormat.Center([]byte(p))
	}
	panic("Center string or []byte only")
}

func Indent(p interface{}) []byte {
	switch p := p.(type) {
	case []byte:
		return DefaultTextFormat.Indent(p)
	case string:
		return DefaultTextFormat.Indent([]byte(p))
	}
	panic("Indent string or []byte only")
}
