package main

import (
	"fmt"
	"os"
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
			// Close all files
			for i := 0; i < maxFiles; i++ {
				if env.file[i] != nil {
					err := env.file[i].Close()
					if err != nil {
						env.raiseError(errorTodo, err.Error())
					}
					env.file[i] = nil
				}
			}
			env.log("OSFIND('Close all files')")
		} else {
			// Close y
			file := env.getFile(y)
			if file != nil {
				err := file.Close()
				if err != nil {
					env.raiseError(errorTodo, err.Error())
				}
				env.file[y-1] = nil
			}
			env.log(fmt.Sprintf("OSFIND('Close file',FILE=%v)", y))
		}
		return
	}

	// Open file
	address := uint16(x) + uint16(y)<<8
	filename := env.mem.getString(address, 0x0d)

	// Find the first free file handle
	handle := -1
	for i := 0; i < maxFiles; i++ {
		if env.file[i] == nil {
			handle = i
			break
		}
	}
	if handle == -1 {
		env.raiseError(190, "Catalogue full")
	} else {
		var err error
		switch a {
		case 0x40: // Open file for input only
			env.file[handle], err = os.Open(filename)
		case 0x80: // Open file for output only
			env.file[handle], err = os.Create(filename)
		case 0xc0: // Open file of update
			env.file[handle], err = os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
		default:
			env.raiseError(errorTodo, fmt.Sprintf("Unknown open mode for OSFIND 0x%02x", a))
			handle = -1
		}
		if err != nil {
			env.raiseError(errorTodo, err.Error())
			handle = -1
		}
	}
	env.cpu.SetAXYP(uint8(handle+1), x, y, p)

	env.log(fmt.Sprintf("OSFIND('Open file',FILE='%s',MODE=0x%02x)=%v",
		filename, a, handle+1))
}
