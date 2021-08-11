package main

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/ivanizag/izapple2/core6502"
)

type environment struct {
	cpu *core6502.State
	mem core6502.Memory
	vdu *vdu
	in  *bufio.Scanner

	// clock, used by OSWORD01 and 02
	referenceTime time.Time

	// timer, used by OSWORD03 and 04
	timer           uint64 // Only 40 bits are used
	lastTimerUpdate time.Time

	// files
	file [maxFiles]*os.File

	// behaviour
	stop bool

	// configuration
	apiLog     bool
	apiLogIO   bool
	panicOnErr bool
}

///////////////////////////
// File handling
///////////////////////////
func (env *environment) getFile(handle uint8) *os.File {
	i := handle - 1
	if i < maxFiles && env.file[i] != nil {
		return env.file[i]
	}

	env.raiseError(222, "Channel")
	return nil
}

///////////////////////////
// Errors and logs
///////////////////////////

func (env *environment) raiseError(code uint8, msg string) {
	/*
		The BBC microcomputer adopts a standard pattern of bytes
		following a BRK instruction, this is:
		A single byte error number
		An error message
		A zero byte to terminate the message

		TODO: set proper error codes
			http://chrisacorns.computinghistory.org.uk/docs/SJR/SJR_HDFSSysMgrManual.pdf
	*/
	env.mem.Poke(errorArea, 0x00 /* BRK opcode */)
	env.mem.Poke(errorArea+1, code)
	env.putStringInMem(errorArea+2, msg, 0, uint8(errorMessageMaxLength))

	env.cpu.SetPC(errorArea)

	env.log(fmt.Sprintf("RAISE(ERR=%02x, '%s')", code, msg))
}

func (env *environment) log(msg string) {
	if env.apiLog {
		fmt.Printf("[[[%s]]]\n", msg)
	}
}

func (env *environment) logIO(msg string) {
	if env.apiLogIO {
		fmt.Printf("[[[%s]]]\n", msg)
	}
}

func (env *environment) notImplemented(feature string) {
	msg := fmt.Sprintf("Not implemented: %s", feature)
	if env.panicOnErr {
		panic(msg)
	}
	env.log(msg)
}

///////////////////////////
// Memory Access
///////////////////////////

func (env *environment) putStringInMem(address uint16, s string, terminator uint8, maxLength uint8) {
	// maxLength not including terminator
	var i int
	for i = 0; i < len(s) && i < int(maxLength); i++ {
		env.mem.Poke(address+uint16(i), s[i])
	}
	env.mem.Poke(address+uint16(i), terminator)
}

func (env *environment) getStringFromMem(address uint16, terminator uint8) string {
	str := ""
	for {
		ch := env.mem.Peek(address)
		//fmt.Printf("{%04x: %02x\n", address, ch)
		if ch == terminator {
			break
		}
		str += string(ch)
		address++
	}
	return str
}

func (env *environment) getMemSlice(address uint16, length uint16) []uint8 {
	slice := make([]uint8, 0, length)
	for i := uint16(0); i < length; i++ {
		slice = append(slice, env.mem.Peek(address+i))
	}
	return slice
}

func (env *environment) storeSliceinMem(address uint16, maxLength uint16, data []uint8) uint16 {
	var i uint16
	for i = 0; i < maxLength && i < uint16(len(data)); i++ {
		env.mem.Poke(address+i, data[i])
	}
	return i
}

func (env *environment) peekWord(address uint16) uint16 {
	return uint16(env.mem.Peek(address)) + uint16(env.mem.Peek(address+1))<<8
}

func (env *environment) pokeWord(address uint16, value uint16) {
	env.mem.Poke(address, uint8(value&0xff))
	env.mem.Poke(address+1, uint8(value>>8))
}

func (env *environment) peeknbytes(address uint16, n int) uint64 {
	ticks := uint64(0)
	for i := n - 1; i >= 0; i-- {
		ticks <<= 8
		ticks += uint64(env.mem.Peek(address + uint16(i)))
	}
	return ticks
}

func (env *environment) pokenbytes(address uint16, n int, value uint64) {
	for i := 0; i < n; i++ {
		env.mem.Poke(address+uint16(i), uint8(value&0xff))
		value >>= 8
	}
}
