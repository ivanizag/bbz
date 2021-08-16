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
	mem *acornMemory
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
	env.storeError(errorArea, code, msg, errorMessageMaxLength)
	env.cpu.SetPC(errorArea)

	env.log(fmt.Sprintf("RAISE(ERR=%02x, '%s')", code, msg))
}

func (env *environment) storeError(address uint16, code uint8, msg string, maxMsgLen int) {
	/*
		The BBC microcomputer adopts a standard pattern of bytes
		following a BRK instruction, this is:
		A single byte error number
		An error message
		A zero byte to terminate the message

		TODO: set proper error codes
			http://chrisacorns.computinghistory.org.uk/docs/SJR/SJR_HDFSSysMgrManual.pdf
	*/
	env.mem.Poke(address, 0x00 /* BRK opcode */)
	env.mem.Poke(address+1, code)
	env.mem.storeString(address+2, msg, 0, uint8(maxMsgLen))

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
