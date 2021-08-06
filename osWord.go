package main

import (
	"fmt"
	"strings"
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
		if !env.in.Scan() {
			env.stop = true
			return
		}
		line := env.in.Text()
		line = strings.ToUpper(line)

		// TODO: check max size
		buffer := env.peekWord(xy)
		maxLength := env.mem.Peek(xy + 2)
		minChar := env.mem.Peek(xy + 3)
		maxChar := env.mem.Peek(xy + 4)
		env.putStringInMem(buffer, line, '\r', maxLength-1)
		pOut := p &^ 1 // Clear carry
		env.cpu.SetAXYP(1, x, uint8(len(line)), pOut)
		env.vdu.mode7Reset()

		env.log(fmt.Sprintf("OSWORD00('read line',BUF=0x%04x,range=%02x-%02x, maxlen=%v)='%s'",
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
		buffer := env.peekWord(xy)
		env.poke5bytes(buffer, uint64(ticks))

		env.log(fmt.Sprintf("OSWORD01('read system clock',BUF=0x%04x)=%v", buffer, ticks&0xff_ffff_ffff))

	case 0x02: // Write system clock
		/*
			This routine may be used to set the system clock to a five byte value contained
			in memory at the address contained in the X and Y registers.
		*/
		buffer := env.peekWord(xy)
		ticks := env.peek5bytes(buffer)
		duration := time.Duration(ticks * 100 * 1000)
		env.referenceTime = time.Now()
		env.referenceTime.Add(duration * -1)

		env.log(fmt.Sprintf("OSWORD02('write system clock',TICKS=%v)", ticks))

	case 0x03: // Read Interval timer
		/*
			In addition to the clock there is an interval timer which is incremented every
			hundredth of a second. The interval is stored in five bytes pointed to by X and Y.
		*/
		duration := time.Since(env.lastTimerUpdate)
		timer := env.timer + uint64(duration.Milliseconds()/10)
		buffer := env.peekWord(xy)
		env.poke5bytes(buffer, uint64(timer))

		env.log(fmt.Sprintf("OSWORD03('read interval timer',BUF=0x%04x)=%v", buffer, timer&0xff_ffff_ffff))

	case 0x04: // Write interval timer
		/*
			On entry X and Y point to five locations which contain the new value to which the
			clock is to be set. The interval timer increments and may cause an event when it
			reaches zero. Thus setting the timer to &FFFFFFFFFE would cause an event
			after two hundredths of a second.
		*/
		buffer := env.peekWord(xy)
		env.timer = env.peek5bytes(buffer)
		env.lastTimerUpdate = time.Now()

		env.log(fmt.Sprintf("OSWORD04('write interval timer',TIMER=%v)", env.timer))

	case 0x05: // Read I/O processor memory
		/*
			A byte of I/O processor memory may be read across the Tube using this call. A 32
			bit address should be contained in memory at the address contained in the X and Y
			registers.
			On exit, The byte read will be contained in location XY+4.
		*/
		address := uint32(env.peekWord(xy)) +
			uint32(env.peekWord(xy+2))<<16
		value := env.mem.Peek((uint16(address)))
		env.mem.Poke(xy+4, value)

		env.log(fmt.Sprintf("OSWORD05('Read I/O processor memory',ADDRESS=0x%08x)=0x%02x",
			address, value))

	default:
		env.notImplemented(fmt.Sprintf("OSWORD%02x", a))
	}
}
