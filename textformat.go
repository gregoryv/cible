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
