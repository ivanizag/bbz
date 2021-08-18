package main

import (
	"bufio"
	"os"
)

type input interface {
	readline() (string, bool)
	readChar() (uint8, bool)
}

type inputSimple struct {
	in *bufio.Scanner
}

func newInputSimple() *inputSimple {
	var i inputSimple
	i.in = bufio.NewScanner(os.Stdin)
	return &i
}

func (i *inputSimple) readline() (string, bool) {
	if !i.in.Scan() {
		return "", true
	}
	line := i.in.Text()
	return line, false
}

func (i *inputSimple) readChar() (uint8, bool) {
	// TODO: capture keystrokes. We will just get the first char of the line
	// and ignore the rest.
	s, stop := i.readline()
	if s == "" {
		return ' ', stop
	} else {
		return s[0], stop
	}
}
