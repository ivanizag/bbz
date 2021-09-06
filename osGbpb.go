package main

import (
	"fmt"
	"io"
)

const (
	cbDataAddress uint16 = 0x1
	cbDataCount   uint16 = 0x5
	cbDataOffset  uint16 = 0x9
)

func execOSGBPB(env *environment) {
	/*
		OSGBPB Read or write a group of bytes.
		This routine transfers a number of bytes to or from an open file.
		It can also be used to transfer filing system information.


			On exit:
		A, X and Y are preserved
		The carry flag is clear in the event of a successful transfer
		C=1 if the transfer could not be completed.

		In the event of a transfer not being completed the parameter block
		contains the following information:
			(a) The number of bytes or names not transferred in the number of bytes to
		transfer field.
			(b) The address field contains the next location of memory due for transfer.
			(c) The sequential pointer field contains the sequential file pointer value
		indicating the next byte in the file due for transfer.

		See: http://beebwiki.mdfs.net/OSGBPB
	*/

	a, x, y, p := env.cpu.GetAXYP()
	if a <= 4 { // R/W multiple bytes of data
		options := []string{
			/*0*/ "no action",
			/*1*/ "write bytes to file at sequential file pointer specified",
			/*2*/ "append bytes to file at current file pointer",
			/*3*/ "read bytes from specified position in file",
			/*4*/ "read bytes from current position in file",
		}
		option := options[a]

		controlBlock := uint16(x) + uint16(y)<<8
		handle := env.mem.Peek(controlBlock)
		address := env.mem.peekDoubleWord(controlBlock + cbDataAddress)
		count := env.mem.peekDoubleWord(controlBlock + cbDataCount)
		offset := env.mem.peekDoubleWord(controlBlock + cbDataOffset)

		error := false
		file := env.getFile(handle)
		if file == nil {
			error = true
			env.log("Invalid handle")
		}

		if !error && (a == 1 || a == 3) { // Use specified pointer
			_, err := file.Seek(int64(offset), io.SeekStart)
			if err != nil {
				error = true
				env.log(err.Error())
			}
		}

		transferred := uint32(0)
		if !error && (a == 1 || a == 2) { // Write
			data := env.mem.peekSlice(uint16(address), uint16(count))
			n, err := file.Write(data)
			if err != nil {
				// Error
				error = true
				env.log(err.Error())
			}
			transferred = uint32(n)
		}

		if !error && (a == 3 || a == 4) { // Read
			data := make([]uint8, count)
			n, err := file.Read(data)
			if err != nil {
				// Error
				error = true
				env.log(err.Error())
			}
			env.mem.pokeSlice(uint16(address), uint16(n), data)
			transferred = uint32(n)
		}

		// Update the file control block
		env.mem.pokeDoubleWord(controlBlock+cbDataAddress, address+transferred)
		env.mem.pokeDoubleWord(controlBlock+cbDataCount, count-transferred)
		env.mem.pokeDoubleWord(controlBlock+cbDataOffset, offset+transferred)

		newP := p
		if count-transferred != 0 {
			newP = newP | 1 // Set carry if not all bytes have been transferred
		} else {
			newP = newP &^ 1 // Clear carry
		}

		env.cpu.SetAXYP(0 /*supported*/, x, y, newP)
		env.log(fmt.Sprintf("OSGBPB('%s',A=%02x,FCB=%04x,FILE=%v,N=%v,ADDRESS=%04x,POS=%v) => (N=%v)",
			option, a, controlBlock, handle, count, address, offset, transferred))
	} else {
		env.notImplemented(fmt.Sprintf("OSGBPB(A=%02x)", a))
	}
}
