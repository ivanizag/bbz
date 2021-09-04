package main

import (
	"fmt"
	"io"
	"time"
)

func execOSBYTE(env *environment) {
	a, x, y, p := env.cpu.GetAXYP()
	newA, newX, newY, newP := a, x, y, p
	option := ""
	isIO := false
	switch a {

	case 0x00:
		option = "Identify Operating System version"
		/*
			Entry parameters:
				X=0 Execute BRK with a message giving the O.S. type
				X<>0 RTS with O.S. type returned in X
			On exit,
				X=0, OS 1.00
				X=1, OS 1.20
		*/
		if x == 0 {
			env.raiseError(errorTodo, "MOS as interpreted by BZZ")
		} else {
			newX = 0
		}

	case 0x02:
		option = "Select input device"
		/*
			On entry, the value in X determines the input device(s), as follows:
				X=0 Keyboard selected, RS423 disabled.
				X=1 RS423 selected, keyboard disabled.
				X=3 Keyboard selected, RS423 enabled (but not selected.)
			On exit, X=0 if previous input was from the keyboard, X=1 if previous input was
			from the RS423. A is preserved, Y an C are undefined.
		*/
		// We do nothing
		newX = 0

	case 0x03:
		option = "Select output device"
		/*
			On entry, the value in X determines the output device to be selected
		*/
		// We do nothing

	case 0x04:
		option = "Enable/disable cursor editing"
		/*
			Entry parameters: X determines editing keys' status, Y=0
		*/
		// We do nothing

	case 0x05:
		option = "Select print destination"
		/*
			Entry parameters: X determines print destination
		*/
		// We do nothing

	case 0x0b:
		option = "Set keyboard auto-repeat delay"
		/*
			Entry parameters: X determines delay before repeating starts
		*/
		// We do nothing

	case 0x0c:
		option = "Set keyboard auto-repeat period"
		/*
			Entry parameters: X determines auto-repeat periodic interval
		*/
		// We do nothing

	case 0x0f:
		option = "Flush specific buffer"
		/*
			Entry parameters: X value selects class of buffer
		*/
		// We do nothing

	case 0x15:
		option = "Flush specific buffer"
		/*
			Entry parameters: X determines the buffer to be cleared
		*/
		// We do nothing

	case 0x72:
		option = "Specify video memory to use on next MODE change"
		/*
			On entry:
				X=0 force use of shadow memory on future MODE changes
				X=1 only use shadow memory if MODE number is > 127
				X=255 never use shadow memory even if MODE > 127 (Solidisk ROM/RAM Extension)
			On exit:
				X=previous setting
		*/
		newX = 1

	case 0x76:
		option = "Reflect keyboard status in keyboard LEDs"
		/*
			This call reflects the keyboard status in the state of the keyboard LEDs, and is
			normally used after the status has been changed by OSBYTE &CA
		*/
		// We do nothing

	case 0x7e:
		option = "Ack detection of an ESC condition"
		/*
			If an ESCAPE condition is detected, all active buffers will be flushed
			and any open EXEC files will be closed. There are no entry conditions;
			on exit, X=255 if the ESCAPE condition existed (X=0 otherwise), A is
			preserved, Y and C are undefined
		*/
		newX = 0

	case 0x7f:
		option = "Check for end-of-file on an opened file"
		/*
			Entry parameters: X contains file handle
			On exit,
			  X<>0 If end-of-file has been reached
			  X=0 If end-of-file has not been reached
		*/
		file := env.getFile(x)
		if file != nil {
			pos, err := file.Seek(0, io.SeekCurrent)
			if err != nil {
				env.raiseError(errorTodo, err.Error())
			} else {
				info, err := file.Stat()
				if err != nil {
					env.raiseError(errorTodo, err.Error())
				} else {
					if pos >= info.Size() {
						newX = 1 // EOF
					} else {
						newX = 0 // Not EOF
					}
				}
			}
		}

	case 0x80:
		option = "Read ADC channel"
		/*
			Entry parameters: X determines action and buffer or channel
			On exit, for input buffers X contains the number of characters in the buffer
			and for output buffers the number of spaces remaining.
		*/
		if x == 0xff { // Keyboard buffer
			newX = 0
		} else {
			env.notImplemented("OSBYTE80 supported only for X=0xff")
		}

	case 0x81:
		/*
			Entry parameters: X and Y specify time limit in centiseconds
			On exit,
				If a character is detected, X=ASCII value of key pressed, Y=0 and C=0.
				If a character is not detected within timeout then Y=&FF and C=1 If Escape
				is pressed then Y=&1B (27) and C=1.
			If called with Y=&FF and a negative INKEY value in X (see appendix C) this call
			performs a keyboard scan. On exit, X and Y contain &FF if the key being scanned
			is pressed and 0 otherwise.
		*/
		//isIO = true

		if x < 0x80 {
			option = "Read key with time limit"
			// We will just wait the time and return that no key was pressed
			timeLimitMs := (uint16(x) + uint16(y)<<8) * 10
			env.logIO(fmt.Sprintf("Sleep(%v ms", timeLimitMs))
			time.Sleep(time.Duration(timeLimitMs) * time.Millisecond)
			newY = 0xff
			newP = newP | 1 // Set carry
		} else {
			option = "Scan keyboard for key press"
			// We will just return that the key was not pressed
			newX = 0
			newY = 0
		}

	case 0x82:
		option = "Read machine high order address"
		/*
			This call provides a 16 bit high order address for filing system
			addresses which require 32 bits. As the BBC microcomputer
			uses 16 bit addresses internally a padding value must be
			provided which associates a given address to that machine.
			On exit, X and Y contain the padding address (X-high, Y-low)
			(This address is &FFFF for the BBC microcomputer I/O
			processor)
		*/
		newX = 0xff
		newY = 0xff

	case 0x83:
		option = "Read bottom of user mem"
		newX = uint8(userMemBottom & 0xff)
		newY = uint8(userMemBottom >> 8)

	case 0x84:
		option = "Read top of user mem"
		newX = uint8(romStartAddress & 0xff)
		newY = uint8(romStartAddress >> 8)

	case 0x85:
		option = "Read top of user mem for mode"
		newX = uint8(romStartAddress & 0xff)
		newY = uint8(romStartAddress >> 8)

	case 0x87:
		option = "Read character at text cursor position"
		/*
			No entry parameters
			On exit,
			X contains character value (0 if char. not recognised)
			Y contains graphics MODE number
		*/
		// Not implemented. Just returns space
		newX = ' '
		newY = env.vdu.mode

	case 0x8b:
		option = "Set filing system options"
		/*
			This call is directly equivalent to *OPT and it controls the computer’s response to
			errors during file operations such as LOAD and SAVE.
			On entry X contains the option number and Y contains the particular selected option
		*/
		// TODO

	case 0x8e:
		option = "Enter language ROM"
		/*
			Entry parameters: X determines which language ROM is entered
			The selected language will be re-entered after a soft BREAK.
			The action of this call is to printout the language name and enter the selected
			language ROM at &8000 with A=1. Locations &FD and &FE in zero page point to the
			copyright message in the ROM. When a Tube is present this call will copy the
			language across to the second processor.
		*/
		env.initLanguage(x)
		newA = 1

	case 0xa0:
		option = "Read VDU variable value"
		/*
			Entry parameters: X contains the number of the number to be read
			On exit, X contains low byte of number and Y contains the high byte
			This call reads locations &300,X and &301,X
		*/
		// Not implemented. Just returns O.
		newX = 0
		newY = 0
		if x == 0x09 {
			//Bottom row / Right hand column
			newX = 31
			newY = 39 // Pascal editor only works with 80 columns
		}

	case 0xca:
		option = "Read/write keyboard status byte"
		/*
			The old value is returned in X. The contents of the next location are returned in Y.
				bit 3: 1 if SHIFT is pressed.
				bit 4: 0 if CAPS LOCK is engaged.
				bit 5: 0 if SHIFT LOCK is engaged.
				bit 6: 1 if CTRL is pressed.
				bit 7: 1 SHIFT enabled, if a LOCK key is engaged then SHIFT reverses the LOCK.
		*/
		// Not implemented. Just returns O.
		newX = 0
		newY = 0

	case 0xda:
		option = "R/W number of items in VDU"
		/*
			Writing 0 to this location can be a useful way of abandoning a
			VDU queue otherwise writing to this location is not
			recommended.
		*/
		if x == 0 || y == 0 {
			env.vdu.clearQueue()
		} else {
			env.notImplemented("OSBYTEda for x or y not zero")
		}

	case 0xea:
		option = "R/W Tube present flag"
		/*
			The value 255 indicates the Tube is present 0 indicates it is not.
			Writing to this value is not recommended.

			We return 0 always
		*/
		newX = 0
		newY = 0

	case 0xfd:
		option = "Read hard/soft break"
		/*
			This location contains a value indicating the type of the last BREAK performed.
				value 0 - soft BREAK
				value 1 - power up reset
				value 2 - hard BREAK
		*/
		newX = 2
		newY = 2

	default:
		// Send to the other ROMS if available.
		env.mem.Poke(zpA, a)
		env.mem.Poke(zpX, x)
		env.mem.Poke(zpY, y)
		newA = serviceOSBYTE
		env.cpu.SetPC(procServiceRoms)
		env.log(fmt.Sprintf("OSBYTE%02x_to_roms(X=0x%02x,Y=0x%02x)", a, x, y))
		// procServiceRoms issues a 254-Bad command if the command is not handled by any ROM
	}

	env.cpu.SetAXYP(newA, newX, newY, newP)
	if option != "" {
		msg := fmt.Sprintf("OSBYTE%02x('%s',X=0x%02x,Y=0x%02x) => (X=0x%02x,Y=0x%02x)", a, option, x, y, newX, newY)
		if isIO {
			env.logIO(msg)
		} else {
			env.log(msg)
		}
	}
}
