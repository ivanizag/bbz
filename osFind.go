package main

import (
	"fmt"
)

func execOSFIND(env *environment) {
	a, x, y, p := env.cpu.GetAXYP()

	/*
		Open or close a file for byte access
		OSFIND is used to open and close files. ‘Opening’ a file declares a file
		requiring byte access to the filing system. ‘Closing’ a file declares that
		byte access is complete. To use OSARGS, OSBGET, OSBPUT, or OSGBPB with a
		file, it must first be opened.

		On entry, The accumulator specifies the operation to be performed:
		If A is zero, a file is to be closed:
			Y contains the handle for the file to be closed. If Y=0, all open files
			are to be closed.

		If A is non zero, a file is to be opened:
			X and Y point to the file name.
			(X = low-byte, Y = high-byte)
			The file name is terminated by carriage return (&0D).

		The accumulator can take the following values:
			&40, a file is to be opened for input only.
			&80, a file is to be opened for output only.
			&C0, a file is to be opened for update (random access).

		When opening a file for output only, an attempt is made to delete the file
		before opening.

		On exit,
		X and Y are preserved.
		A is preserved on closing, and on opening contains the file handle assigned
		to the file. If A=0 on exit, the file could not be opened.
	*/
	if a == 0 {
		// Close file
		if y == 0 {
			for i := uint8(0); i < maxFiles; i++ {
				env.closeFile(i + 1)
			}
			env.log("OSFIND('Close all files')")
		} else {
			env.closeFile(y)
			env.log(fmt.Sprintf("OSFIND('Close file',FILE=%v)", y))
		}
		return
	}

	// Open file
	address := uint16(x) + uint16(y)<<8
	filename := env.mem.peekString(address, 0x0d)
	file := env.openFile(filename, a)
	env.cpu.SetAXYP(file, x, y, p)

	env.log(fmt.Sprintf("OSFIND('Open file',FILE='%s',MODE=0x%02x)=%v",
		filename, a, file))
}
