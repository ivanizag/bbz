package main

import (
	"fmt"
	"os"
)

func (env *environment) getFile(handle uint8) *os.File {
	i := handle - 1
	if i < maxFiles && env.file[i] != nil {
		return env.file[i]
	}

	env.raiseError(222, "Channel")
	return nil
}

func (env *environment) openFile(filename string, mode uint8) uint8 {
	// Find the first free file handle
	i := 0
	for ; i < int(maxFiles); i++ {
		if env.file[i] == nil {
			break
		}
	}
	if i == int(maxFiles) {
		env.raiseError(190, "Catalogue full")
		return 0
	}

	var err error
	switch mode {
	case 0x40: // Open file for input only
		//env.file[i], err = os.Open(filename)
		env.file[i], err = os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0644)
	case 0x80: // Open file for output only
		env.file[i], err = os.Create(filename)
	case 0xc0: // Open file foe update
		env.file[i], err = os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	default:
		env.raiseError(errorTodo, fmt.Sprintf("Unknown open mode for OSFIND 0x%02x", mode))
		i = -1
	}
	if err != nil {
		env.raiseError(errorTodo, err.Error())
		i = -1
	}

	return uint8(i + 1)
}

func (env *environment) closeFile(handle uint8) {
	if handle == 0 {
		return
	}
	i := handle - 1
	if env.file[i] != nil {
		err := env.file[i].Close()
		if err != nil {
			env.raiseError(errorTodo, err.Error())
		}
		env.file[i] = nil
	}
}

func (env *environment) writeSpool(s string) {
	charDest := env.mem.Peek(charDestinations)
	if charDest&0x10 != 0 {
		// Spooled output is disabled
		return
	}

	spoolHandle := env.mem.Peek(spoolFileHandle)
	if spoolHandle == 0 {
		// No spool file defined
		return
	}
	file := env.getFile(spoolHandle)

	fmt.Fprintf(file, "%s", s)
}
