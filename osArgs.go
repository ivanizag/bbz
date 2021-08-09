package main

import (
	"fmt"
	"io"
)

func execOSARGS(env *environment) {
	a, x, y, _ := env.cpu.GetAXYP()

	/*
		Call address &FFDA Indirected through &214
		This routine reads or writes an open file's attributes.
		On entry, X points to a four byte zero page control block.
		Y contains the file handle as provided by OSFIND, or zero.
		The accumulator contains a number specifying the action required.
	*/

	if y == 0 {
		// Operations without file handle
		switch a {
		case 0: // Returns the current filing system in A

			filingSystem := 9 // Host filing system
			env.log(fmt.Sprintf("OSARGS('Get filing system',A=%02x,Y=%02x)= %v", a, y, filingSystem))

		case 0xff: // Update all files onto the media
			// Do nothing.
			env.log("OSARGS('Update all files onto the media')")

		default:
			env.notImplemented(fmt.Sprintf("OSARGS(A=%02x,Y=%02x)", a, y))
		}

		return
	}

	file := env.getFile(y)
	if file == nil {
		env.log(fmt.Sprintf("OSARGS(A=%02x,FILE=%v)='bad handler'", a, y))
	}

	switch a {
	case 0: // Read sequential pointer of file (BASIC PTR#)
		pos, err := file.Seek(0, io.SeekCurrent)
		if err != nil {
			env.raiseError(errorTodo, err.Error())
		} else {
			env.pokenbytes(uint16(x), 4, uint64(pos))
		}
		env.log(fmt.Sprintf("OSARGS('Get PTR#',FILE=%v)=%v", y, pos))

	case 1: // Write sequential pointer of file
		pos := int64(env.peeknbytes(uint16(x), 4))
		_, err := file.Seek(pos, io.SeekStart)
		if err != nil {
			env.raiseError(errorTodo, err.Error())
		}
		env.log(fmt.Sprintf("OSARGS('Set PTR#',FILE=%v,PTR=%v)", y, pos))

	case 2: // Read length of file (BASIC EXT#)
		info, err := file.Stat()
		if err != nil {
			env.raiseError(errorTodo, err.Error())
		} else {
			env.pokenbytes(uint16(x), 4, uint64(info.Size()))
		}
		env.log(fmt.Sprintf("OSARGS('Get EXT#',FILE=%v)=%v", y, info.Size()))

	case 0xff: // Update this file to media
		// Do nothing.
		env.log(fmt.Sprintf("OSARGS('Update file to media',FILE=%v)", y))

	default:
		env.notImplemented(fmt.Sprintf("OSARGS(A=%02x,FILE=%v)", a, y))
	}
}
