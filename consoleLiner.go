package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/peterh/liner"
)

const historyFilename = ".bbzhistory"

type consoleLiner struct {
	liner  *liner.State
	prompt string
	env    *environment
}

func newConsoleLiner(env *environment) *consoleLiner {
	var c consoleLiner

	c.liner = liner.NewLiner()
	c.env = env
	c.liner.SetCtrlCAborts(true)
	if f, err := os.Open(historyFilename); err == nil {
		c.liner.ReadHistory(f)
		f.Close()
	}

	return &c
}

func (c *consoleLiner) close() {

	if f, err := os.Create(historyFilename); err == nil {
		c.liner.WriteHistory(f)
		f.Close()
	}

	c.liner.Close()
}

func (c *consoleLiner) readline() (string, bool) {
	fmt.Printf("\r")
	line, err := c.liner.Prompt(c.prompt)
	if errors.Is(err, liner.ErrInvalidPrompt) {
		fmt.Println()
		line, err = c.liner.Prompt("")
	}
	c.prompt = ""
	if errors.Is(err, liner.ErrPromptAborted) {
		c.env.escape()
		return "", false
	}
	if errors.Is(err, io.EOF) {
		return "", true
	}
	if err != nil {
		panic(err)
	}
	if line != "" {
		c.liner.AppendHistory(line)
	}
	return line, false
}

func (c *consoleLiner) readChar() (uint8, bool) {
	// TODO: capture keystrokes. We will just get the first char of the line
	// and ignore the rest.
	s, stop := c.readline()
	if s == "" {
		return ' ', stop
	} else {
		return s[0], stop
	}
}

func (c *consoleLiner) write(s string) {
	if strings.HasSuffix(s, "\n") || strings.HasSuffix(s, "\r") {
		c.prompt = ""
	} else {
		c.prompt += s
	}
	fmt.Print(s)
}
