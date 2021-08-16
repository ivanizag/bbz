package main

import (
	"fmt"
	"io/ioutil"
)

func execOSFILE(env *environment) {
	a, x, y, _ := env.cpu.GetAXYP()

	// See: http://beebwiki.mdfs.net/OSFILE
	controlBlock := uint16(x) + uint16(y)<<8
	filenameAddress := env.mem.peekWord(controlBlock)
	loadAddress := env.mem.peekWord(controlBlock + 0x2)
	executionAddress := env.mem.peekWord(controlBlock + 0x6)
	startAddress := env.mem.peekWord(controlBlock + 0xa)
	endAddress := env.mem.peekWord(controlBlock + 0xe)

	filename := env.mem.getString(filenameAddress, 0x0d)
	filesize := endAddress - startAddress

	switch a {
	case 0:
		/*
			Save a block of memory as a file using the
			information provided in the parameter block.
		*/
		data := env.mem.getSlice(startAddress, filesize)
		err := ioutil.WriteFile(filename, data, 0644)
		if err != nil {
			env.raiseError(errorTodo, err.Error())
		}
	case 0xff: // Load file into memory
		/*
			Load the named file, the address to which the file is
			loaded being determined by the lowest byte of the
			execution address in the control block (XY+6). If
			this byte is zero, the address given in the control
			block is used, otherwise the file's own load address
			is used.
		*/
		useLoadAddress := (executionAddress & 0xff) == 0
		if !useLoadAddress {
			env.notImplemented("Loading files on their own load address")
		}
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			env.raiseError(errorTodo, err.Error())
		}
		// NOTE: There is no maxLength?
		env.mem.storeSlice(loadAddress, uint16(len(data)), data)
		filesize = uint16(len(data))
	}

	env.mem.pokeWord(controlBlock+0xa, filesize)
	env.mem.pokeWord(controlBlock+0x3, 0x33 /*attributes*/)

	env.log(fmt.Sprintf("OSFILE(A=%02x,FCB=%04x,FILE=%s,SIZE=%v)", a, controlBlock, filename, filesize))

}
