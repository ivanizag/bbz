package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"time"
	"unicode"

	"github.com/ivanizag/izapple2/core6502"
)

func RunMOSEnvironment(romFilename string, cpuLog bool, apiLog bool, apiLogIO bool, panicOnErr bool) {
	// Prepare environment
	var env environment
	env.in = bufio.NewScanner(os.Stdin)
	env.referenceTime = time.Now()
	env.timer = 0
	env.lastTimerUpdate = time.Now()
	env.mem = new(core6502.FlatMemory)
	env.cpu = core6502.NewNMOS6502(env.mem)
	env.cpu.SetTrace(cpuLog)
	env.vdu = newVdu()

	env.loadRom(romFilename)
	env.loadMos()

	env.apiLog = apiLog
	env.apiLogIO = apiLogIO
	env.panicOnErr = panicOnErr

	// Execute
	for !env.stop {
		env.cpu.ExecuteInstruction()

		pc, sp := env.cpu.GetPCAndSP()
		if pc >= mosVectors {
			a, x, y, p := env.cpu.GetAXYP()

			// Intercept MOS API calls.
			switch pc {
			case 0xff0f: // OXBRK
				// The selected ROM has not defined a custom BRKV
				panic("Unhandled BRK")

			case 0xff15: // 6502 BRK handler
				/*
					When the 6512 encounters a BRK instruction the operating system places
					the address following the BRK instruction in locations &FD and &FE. Thus
					these locations point to the ‘fault number’. The operating system then
					indirects via location &202.

					The BBC micro uses the BRK instruction to change program flow and cause an error.
					The microprocessor saves the addr+2 of the first BRK on the stack high byte first,
					and the processor's status register is copied onto the stack too.
					Next, the OS saves the accumulator at &FC, pulls the stack (the processor's status
					register) then pushes the processor's status register back onto the stack leaving
					a copy in the accumulator. This is ANDed with 16 to isolate the BRK bit. If it was
					not a BRK interrupt then the OS jumps through IRQ1V. If bit 4 was set then the addr+2
					is read in, has one subtracted from it and then this is stored at &FD and &FE.
					Provided the error was set up in the above format this address now points to the error
					number.
					Interrupts are re-enabled, and the operating system jumps through BRKV (which the
					language has previously set up to point to its error handler).
				*/
				pStacked := env.mem.Peek(0x100 + uint16(sp+1))
				address := env.peekWord(0x100+uint16(sp+2)) - 1
				faultNumber := env.mem.Peek(address)
				faultMessage := address + 1
				faultString := env.getStringFromMem(faultMessage, 0)

				env.mem.Poke(zpAccumulator, a)
				env.pokeWord(zpErrorPointer, address)
				env.cpu.SetAXYP(pStacked&0x10, x, y, p)
				brkv := env.peekWord(mosVectorBrk)
				env.cpu.SetPC(brkv)

				env.log(fmt.Sprintf("BREAK(ERR=%02x, '%s')", faultNumber, faultString))

			case 0xffce: // OSFIND
				execOSFIND(&env)

			case 0xffd4: // OSBPUT
				/*
					Write a single byte to an open file
					On entry, Y contains the file handle, as provided by OSFIND. A contains the
					byte to be written. The byte is placed at the point in the file designated by
					the sequential pointer.
				*/
				file := env.getFile(y)
				if file != nil {
					buf := []uint8{a}
					_, err := file.Write(buf)
					if err != nil {
						env.raiseError(errorTodo, err.Error())
					}
				}
				env.log(fmt.Sprintf("OSBPUT(FILE=%v,VAL=0x%02x)", y, a))

			case 0xffd7: // OSBGET
				/*
					Get one byte from an open file
					This routine reads a single byte from a file.
					On entry, Y contains the file handle, as provided by OSFIND. The byte is
					obtained from the point in the file designated by the sequential pointer.
					On exit, A contains the byte read. C is set if the end of the file has been
					reached, and indicates that the byte obtained is invalid.
				*/
				value := uint8(0)
				eof := false
				file := env.getFile(y)
				if file != nil {
					buf := make([]uint8, 1)
					_, err := file.Read(buf)
					if err == io.EOF {
						// EOF, set C
						eof = true
						env.cpu.SetAXYP(a, x, y, p|1)
					} else if err != nil {
						env.raiseError(errorTodo, err.Error())
					} else {
						// Valid, clear C
						value = buf[0]
						env.cpu.SetAXYP(buf[0], x, y, p&0xfe)
					}
				}
				env.log(fmt.Sprintf("OSBGET(FILE=%v)=0x%02x,EOF=%v", y, value, eof))

			case 0xffda: // OSARGSexecOSARGS
				execOSARGS(&env)

			case 0xffdd: // OSFILE: Load or save a complete file. BPUG page 446
				execOSFILE(&env)

			case 0xffe0: // OSRDCH
				/*
					This routine reads a character from the currently selected input
					stream and returns the character read in the accumulator.
					After an OSRDCH call: C=0 indicates that a valid character has
					been read; C=1 flags an error condition, A contains an error number.
				*/
				if !env.in.Scan() {
					return
				}
				line := env.in.Text()
				// TODO: capture keystrokes. We will just get the first chat of the line
				// and ignore the rest.
				ch := line[0]
				pOut := p &^ 1 // Clear carry
				env.cpu.SetAXYP(ch, x, y, pOut)

				env.log(fmt.Sprintf("OSRDCH()=0x%02x", ch))

			case 0xffe3: // OSASCI
				/*
					This routine performs an OSWRCH call with the accumulator
					contents unless called with accumulator contents of &0D (13)
					when an OSNEWL call is performed.
				*/
				env.vdu.writeAscii(a)
				env.logIO(fmt.Sprintf("OSASXCI(0x%02x, '%v')", a, printableChar(a)))

			case 0xffe7: // OSNEWL
				/*
					This call issues an LF CR to the currently selected output stream.
				*/
				env.vdu.writeNewline()
				env.logIO("OSNEWL()")

			case 0xffee: // OSWRCH
				/*
					This call writes the character in A to the currently selected output stream.
				*/
				env.vdu.write(a)
				env.logIO(fmt.Sprintf("OSWRCH(0x%02x, '%v')", a, printableChar(a)))

			case 0xfff1: // OSWORD
				execOSWORD(&env)
			case 0xfff4: // OSBYTE
				execOSBYTE(&env)
			case 0xfff7: // OSCLI
				execOSCLI(&env)

			default:
				env.notImplemented(fmt.Sprintf("MOS call to 0x%04x", pc))
			}
		}
	}
}

func printableChar(i uint8) string {
	ch := string(i)
	if !unicode.IsGraphic([]rune(ch)[0]) {
		ch = ""
	}
	return ch
}
