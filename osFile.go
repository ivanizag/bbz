package main

import (
	"fmt"
	"io/ioutil"
)

func execOSFILE(env *environment) {
	a, x, y, _ := env.cpu.GetAXYP()

	// See: http://beebwiki.mdfs.net/OSFILE
	controlBlock := uint16(x) + uint16(y)<<8
	filenameAddress := env.peekWord(controlBlock)
	loadAddress := env.peekWord(controlBlock + 0x2)
	executionAddress := env.peekWord(controlBlock + 0x6)
	startAddress := env.peekWord(controlBlock + 0xa)
	endAddress := env.peekWord(controlBlock + 0xe)

	filename := env.getStringFromMem(filenameAddress, 0x0d)
	filesize := endAddress - startAddress

	switch a {
	case 0:
		/*
			Save a block of memory as a file using the
			information provided in the parameter block.
		*/
		data := env.getMemSlice(startAddress, filesize)
		err := ioutil.WriteFile(filename, data, 0644)
		if err != nil {
			panic(err)
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
			panic(err)
		}
		// NOTE: There is no maxLength?
		env.storeSliceinMem(loadAddress, uint16(len(data)), data)
		filesize = uint16(len(data))
	}

	env.pokeWord(controlBlock+0xa, filesize)
	env.pokeWord(controlBlock+0x3, 0x33 /*attributes*/)

	env.log(fmt.Sprintf("OSFILE(A=%02x,FCB=%04x,FILE=%s,SIZE=%v)", a, controlBlock, filename, filesize))

}
