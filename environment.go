package main

import (
	"fmt"
	"os"
	"time"

	"github.com/ivanizag/izapple2/core6502"
)

type environment struct {
	cpu *core6502.State
	mem *acornMemory
	vdu *vdu
	con console

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

func newEnvironment(cpuLog bool, apiLog bool, apiLogIO bool, memLog bool, panicOnErr bool, readline bool) *environment {
	var env environment
	if readline {
		env.con = newConsoleLiner()
	} else {
		env.con = newConsoleSimple()
	}
	env.referenceTime = time.Now()
	env.timer = 0
	env.lastTimerUpdate = time.Now()
	env.mem = newAcornMemory(memLog)
	env.cpu = core6502.NewNMOS6502(env.mem)
	//env.cpu = core6502.NewCMOS65c02(env.mem)
	env.cpu.SetTrace(cpuLog)
	env.vdu = newVdu(env.con)
	env.apiLog = apiLog
	env.apiLogIO = apiLogIO
	env.panicOnErr = panicOnErr
	return &env
}

func (env *environment) close() {
	env.con.close()
}

func (env *environment) getFile(handle uint8) *os.File {
	i := handle - 1
	if i < maxFiles && env.file[i] != nil {
		return env.file[i]
	}

	env.raiseError(222, "Channel")
	return nil
}

func (env *environment) initUpperLanguage() {
	for slot := 0xf; slot >= 0; slot-- {
		romType := env.mem.data[romTypeTable+uint16(slot)]
		if romType&0x40 != 0 {
			env.initLanguage(uint8(slot))
			return
		}
	}

	panic("There is no language ROM available to boot")
}

func (env *environment) initLanguage(slot uint8) {
	env.mem.Poke(zpROMSelect, slot)
	env.mem.Poke(sheilaRomLatch, slot)

	/*
		Next, the MOS will set the error point at &FD/&FE to point at the version string (or copyright
		message if no version string is present).
	*/
	copyrightAddress := 0x8000 + 1 + uint16(env.mem.Peek(romCopyrightOffsetPointer))
	env.mem.pokeWord(zpErrorPointer, copyrightAddress)
	/*
		The MOS also automatically prints the ROM's title string (&8009) so that the user is acknowledged.
	*/
	language := env.mem.getString(romTitleString, 0)
	fmt.Printf("%s\n", language)

	_, x, y, p := env.cpu.GetAXYP()
	env.cpu.SetAXYP(1, x, y, p)
	env.cpu.SetPC(romStartAddress)
}

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
