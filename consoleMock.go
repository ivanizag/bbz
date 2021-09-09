package main

type consoleMock struct {
	linesIn []string
	lineIn  int
	output  string
	env     *environment
}

func newConsoleMock(env *environment, linesIn []string) *consoleMock {
	var c consoleMock
	c.linesIn = linesIn
	c.output = ""
	c.env = env
	return &c
}

func (c *consoleMock) readline() (string, bool) {
	if c.lineIn >= len(c.linesIn) {
		c.env.stop = true
		return "", true
	}
	line := c.linesIn[c.lineIn]
	c.env.writeSpool(line)
	c.env.writeSpool("\n")
	c.lineIn++
	return line, false
}

func (c *consoleMock) readChar() (uint8, bool) {
	s, stop := c.readline()
	if s == "" {
		return ' ', stop
	} else {
		return s[0], stop
	}
}

func (c *consoleMock) write(s string) {
	c.output += s
	c.env.writeSpool(s)
}

func (c *consoleMock) close() {}
