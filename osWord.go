package bbz

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
		env.in.Scan()
		line := env.in.Text()
		line = strings.ToUpper(line)

		// TODO: check max size
		buffer := env.peekWord(xy)
		maxLength := env.mem.Peek(xy + 2)
		minChar := env.mem.Peek(xy + 3)
		maxChar := env.mem.Peek(xy + 4)
		env.putStringInMem(buffer, line+"\r", maxLength)
		pOut := p &^ 1 // Clear carry
		env.cpu.SetAXYP(1, x, uint8(len(line)), pOut)

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
		ticksLog := ticks & 0xff_ffff_ffff // 5 bytes
		buffer := env.peekWord(xy)
		for i := uint16(0); i < 5; i++ {
			env.mem.Poke(buffer+i, uint8(ticks&0xff))
			ticks >>= 8
		}

		env.log(fmt.Sprintf("OSWORD01('read system clock',BUF=0x%04x)=%v", buffer, ticksLog))

	case 0x02: // Write system clock
		/*
			This routine may be used to set the system clock to a five byte value contained
			in memory at the address contained in the X and Y registers.
		*/
		buffer := env.peekWord(xy)
		ticks := uint64(0)
		for i := uint16(0); i < 5; i++ {
			ticks <<= 8
			ticks += uint64(env.mem.Peek(buffer + i))
		}
		duration := time.Duration(ticks * 100 * 1000)
		env.referenceTime = time.Now()
		env.referenceTime.Add(duration * -1)

		env.log(fmt.Sprintf("OSWORD02('write system clock',TICKS=%v)", ticks))

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
