package main

import "fmt"

func execOSBYTE(env *environment) {
	a, x, y, p := env.cpu.GetAXYP()
	newA, newX, newY, newP := a, x, y, p
	option := ""
	switch a {
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

	case 0x7e:
		option = "Ack detection of an ESC condition"
		/*
			If an ESCAPE condition is detected, all active buffers will be flushed
			and any open EXEC files will be closed. There are no entry conditions;
			on exit, X=255 if the ESCAPE condition existed (X=0 otherwise), A is
			preserved, Y and C are undefined
		*/
		newX = 0

	case 0x80:
		option = "Read ADC channel"
		if x == 0xff {
			newX = 0
		} else {
			env.notImplemented("OSBYTE80 supported only for X=ff")
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
		newX = uint8(langRomStart & 0xff)
		newY = uint8(langRomStart >> 8)

	case 0x85:
		option = "Read top of user mem for mode"
		newX = uint8(langRomStart & 0xff)
		newY = uint8(langRomStart >> 8)

	case 0x8b:
		option = "xx"
		/*
			This call is directly equivalent to *OPT and it controls the computer’s response to
			errors during file operations such as LOAD and SAVE.
			On entry X contains the option number and Y contains the particular selected option
		*/
		// TODO

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

	default:
		env.notImplemented(fmt.Sprintf("OSBYTE%02x", a))
	}

	env.cpu.SetAXYP(newA, newX, newY, newP)
	if option != "" {
		env.log(fmt.Sprintf("OSBYTE%02x('%s',X=0x%02x,Y=0x%02x) => (X=0x%02x,Y=0x%02x)", a, option, x, y, newX, newY))
	}
}
