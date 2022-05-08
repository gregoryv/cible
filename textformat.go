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
	indent := f.Prefix(p)

	scanner := bufio.NewScanner(bytes.NewReader(p))
	var buf bytes.Buffer
	for scanner.Scan() {
		line := scanner.Text()
		buf.Write(indent)
		buf.WriteString(line)
		buf.WriteString("\n")
	}
	return bytes.TrimRight(buf.Bytes(), "\n")
}

func (f *TextFormat) Prefix(p []byte) []byte {
	scanner := bufio.NewScanner(bytes.NewReader(p))
	var size int
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) > size {
			size = len(line)
		}
	}
	margin := (f.cols - size) / 2
	if margin > 0 {
		return []byte(strings.Repeat(" ", margin))
	}
	return []byte{}
}

func (f *TextFormat) Center(p []byte) []byte {
	return CenterIn(p, f.cols)
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

func CenterIn(p []byte, width int) []byte {
	scanner := bufio.NewScanner(bytes.NewReader(p))
	var buf bytes.Buffer
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < width {
			prefix := strings.Repeat(" ", (width-len(line))/2)
			buf.WriteString(prefix)
			buf.WriteString(line)
		}
		buf.WriteString("\n")
	}
	return bytes.TrimRight(buf.Bytes(), "\n")
}
