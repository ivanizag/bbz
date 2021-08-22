package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

func execOSFILE(env *environment) {
	a, x, y, p := env.cpu.GetAXYP()

	// See: http://beebwiki.mdfs.net/OSFILE
	controlBlock := uint16(x) + uint16(y)<<8
	filenameAddress := env.mem.peekWord(controlBlock)
	loadAddress := env.mem.peeknbytes(controlBlock+0x2, 4)
	executionAddress := env.mem.peeknbytes(controlBlock+0x6, 4)
	startAddress := env.mem.peeknbytes(controlBlock+0xa, 4)
	endAddress := env.mem.peeknbytes(controlBlock+0xe, 4)

	filename := env.mem.getString(filenameAddress, 0x0d)
	filesize := endAddress - startAddress

	newA := uint8(0) // Nothing found
	option := ""
	switch a {
	case 0:
		option = "Save file"
		/*
			Save a block of memory as a file using the
			information provided in the parameter block.
		*/
		data := env.mem.getSlice(uint16(startAddress), uint16(filesize))
		err := ioutil.WriteFile(filename, data, 0644)
		if err != nil {
			env.raiseError(errorTodo, err.Error())
		}
		newA = 1 // File found
	case 5:
		option = "File info"
		/*
			Read a file’s catalogue information, with the file
			type returned in the accumulator. The information
			is written to the parameter block.
		*/
		info, err := os.Stat(filename)

		if err == nil {
			filesize = uint64(info.Size())
			if info.IsDir() {
				newA = 2 // Directory found
			} else {
				newA = 1 // File found
			}
		}

	case 0xff: // Load file into memory
		option = "Load file"
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
		if err == nil {
			// NOTE: There is no maxLength?
			env.mem.storeSlice(uint16(loadAddress), uint16(len(data)), data)
			filesize = uint64(len(data))
			newA = 1 // File found
		}
	default:
		env.notImplemented(fmt.Sprintf("OSFILE(A=%02x)", a))
	}

	env.mem.pokenbytes(controlBlock+0xa, 4, filesize)
	env.mem.pokeWord(controlBlock+0x3, 0x00 /*attributes*/)

	env.cpu.SetAXYP(newA, x, y, p)
	env.log(fmt.Sprintf("OSFILE('%s',A=%02x,FCB=%04x,FILE=%s,SIZE=%v) => (A=%v,SIZE=%v)",
		option, a, controlBlock, filename, filesize, newA, filesize))

}
