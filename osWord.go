package main

import (
	"fmt"
	"time"
)

func execOSWORD(env *environment) {
	a, x, y, p := env.cpu.GetAXYP()
	xy := uint16(x) + uint16(y)<<8

	switch a {
	case 0x00: // Read line
		/*
			This routine accepts characters from the current input stream and
			places them at a specified location in memory. During input the delete
			code (ASCII 127) deletes the last character entered, and CTRL U (ASCII
			21)  deletes  the  entire line.  The routine ends if RETURN is entered
			(ASCII 13) or an ESCAPE condition occurs.
			On exit C=0 indicates that RETURN (CR; ASCII code 13 or &D) ended the
			line. C not equal to zero indicates that an escape condition terminated
			entry. Y is set to the length of the line, excluding the CR if C=0.
		*/
		line, stop := env.con.readline()
		if stop {
			env.stop = true
			return
		}

		// TODO: check max size
		buffer := env.mem.peekWord(xy)
		maxLength := env.mem.Peek(xy + 2)
		minChar := env.mem.Peek(xy + 3)
		maxChar := env.mem.Peek(xy + 4)
		env.mem.storeString(buffer, line, '\r', maxLength-1)
		pOut := p &^ 1 // Clear carry
		env.cpu.SetAXYP(1, x, uint8(len(line)), pOut)
		env.vdu.mode7Reset()

		env.log(fmt.Sprintf("OSWORD00('read line',BUF=0x%04x,range=%02x-%02x, maxlen=%v) => '%s'",
			buffer, minChar, maxChar, maxLength, line))

	case 0x01: // Read system clock
		/*
			This routine may be used to read the system clock (used for the TIME function
			in BASIC). The five byte clock value is written to the address contained in the
			X and Y registers. This clock is incremented every hundredth of a second and is
			set to 0 by a hard BREAK.
		*/
		duration := time.Since(env.referenceTime)
		ticks := duration.Milliseconds() / 10
		env.mem.pokenbytes(xy, 5, uint64(ticks))

		env.log(fmt.Sprintf("OSWORD01('read system clock',BUF=0x%04x) => %v", xy, ticks&0xff_ffff_ffff))

	case 0x02: // Write system clock
		/*
			This routine may be used to set the system clock to a five byte value contained
			in memory at the address contained in the X and Y registers.
		*/
		ticks := env.mem.peeknbytes(xy, 5)
		duration := time.Duration(ticks * 10 * uint64(time.Millisecond))
		env.referenceTime = time.Now()
		env.referenceTime = env.referenceTime.Add(duration * -1)

		env.log(fmt.Sprintf("OSWORD02('write system clock',TICKS=%v)", ticks))

	case 0x03: // Read Interval timer
		/*
			In addition to the clock there is an interval timer which is incremented every
			hundredth of a second. The interval is stored in five bytes pointed to by X and Y.
		*/
		duration := time.Since(env.lastTimerUpdate)
		timer := env.timer + uint64(duration.Milliseconds()/10)
		env.mem.pokenbytes(xy, 5, uint64(timer))

		env.log(fmt.Sprintf("OSWORD03('read interval timer',BUF=0x%04x) => %v", xy, timer&0xff_ffff_ffff))

	case 0x04: // Write interval timer
		/*
			On entry X and Y point to five locations which contain the new value to which the
			clock is to be set. The interval timer increments and may cause an event when it
			reaches zero. Thus setting the timer to &FFFFFFFFFE would cause an event
			after two hundredths of a second.
		*/
		env.timer = env.mem.peeknbytes(xy, 5)
		env.lastTimerUpdate = time.Now()

		env.log(fmt.Sprintf("OSWORD04('write interval timer',TIMER=%v)", env.timer))

	case 0x05: // Read I/O processor memory
		/*
			A byte of I/O processor memory may be read across the Tube using this call. A 32
			bit address should be contained in memory at the address contained in the X and Y
			registers.
			On exit, The byte read will be contained in location XY+4.
		*/
		address := uint32(env.mem.peekWord(xy)) +
			uint32(env.mem.peekWord(xy+2))<<16
		value := env.mem.Peek((uint16(address)))
		env.mem.Poke(xy+4, value)

		env.logIO(fmt.Sprintf("OSWORD05('Read I/O processor memory',ADDRESS=0x%08x) => 0x%02x",
			address, value))

	case 0x06: // Write i/O processor memory
		/*
			This call permits I/O processor memory to be written across the Tube. A 32 bit
			address is contained in the parameter block addressed by the X and Y registers
			and the byte to be written should be placed in XY+4.
		*/
		address := uint32(env.mem.peekWord(xy)) +
			uint32(env.mem.peekWord(xy+2))<<16
		value := env.mem.Peek(xy + 4)
		env.mem.Poke(uint16(address), value)

		env.log(fmt.Sprintf("OSWORD06('Write I/O processor memory',ADDRESS=0x%08x,VAL=0x%02x)",
			address, value))

	case 0x07: // Sound command
		/*
			This routine takes an 8 byte parameter block addressed by the X and Y registers. The 8
			bytes of the parameter block may be considered as the four parameters used for the SOUND
			command in BASIC.
		*/
		channel := env.mem.peekWord(xy)
		amplitude := int8(env.mem.peekWord(xy + 2))
		pitch := env.mem.peekWord(xy + 4)
		duration := env.mem.peekWord(xy + 6)
		// TODO: play sound

		env.log(fmt.Sprintf("OSWORD07('Sound command',CHAN=%v,AMPL=%v,PITCH=%v,DUR=%v)",
			channel, amplitude, pitch, duration))

	case 0x08: // Define an envelope
		/*
			The ENVELOPE parameter block should contain 14 bytes of data which correspond to the 14
			parameters described in the ENVELOPE command. This call should be entered with the
			parameter block address contained in the X and Y registers.
		*/
		number := env.mem.Peek(xy)
		// TODO: define envelope

		env.log(fmt.Sprintf("OSWORD08('Define envelope',NUMBER=%v)", number))

	default:
		env.notImplemented(fmt.Sprintf("OSWORD%02x", a))
	}
}
