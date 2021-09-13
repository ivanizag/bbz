package main

import (
	"bufio"
	"fmt"
	"os"
)

type console interface {
	readline() (string, bool)
	readChar() (uint8, bool)
	write(string)
	writef(string, ...interface{})
	close()
}

type consoleSimple struct {
	in  *bufio.Scanner
	env *environment
}

func newConsoleSimple(env *environment) *consoleSimple {
	var c consoleSimple
	c.in = bufio.NewScanner(os.Stdin)
	c.env = env
	return &c
}

func (c *consoleSimple) readline() (string, bool) {
	if !c.in.Scan() {
		return "", true
	}
	line := c.in.Text()
	c.env.writeSpool(line)
	c.env.writeSpool("\n")
	return line, false
}

func (c *consoleSimple) readChar() (uint8, bool) {
	// TODO: capture keystrokes. We will just get the first char of the line
	// and ignore the rest.
	s, stop := c.readline()
	if s == "" {
		return ' ', stop
	} else {
		return s[0], stop
	}
}

func (c *consoleSimple) write(s string) {
	fmt.Print(s)
	c.env.writeSpool(s)
}

func (c *consoleSimple) writef(format string, a ...interface{}) {
	s := fmt.Sprintf(format, a...)
	c.write(s)
}

func (c *consoleSimple) close() {}
