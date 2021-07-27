package bbz

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/ivanizag/izapple2/core6502"
)

const (
	zpErrorPointer uint16 = 0x00fd
	mosVectorBrk   uint16 = 0x0202

	userMemBottom uint16 = 0x0e00
	langRomStart  uint16 = 0x8000

	mosVectors      uint16 = 0xff00
	breakEntryPoint uint16 = 0xffb0 // Invented as there is not an address defined
	vectorReset     uint16 = 0xfffc
	vectorBreak     uint16 = 0xfffe
)

func RunMOSEnvironment(romFilename string, cpuLog bool, apiLog bool, apiLogIO bool) {
	console := bufio.NewScanner(os.Stdin)
	referenceTime := time.Now() // Use by OSWORD01

	// Prepare cpu and memory
	memory := new(core6502.FlatMemory)
	cpu := core6502.NewNMOS6502(memory)
	cpu.SetTrace(cpuLog)

	// Load ROM
	romFile, err := os.Open(romFilename)
	if err != nil {
		panic(err)
	}
	data, err := ioutil.ReadAll(romFile)
	if err != nil {
		panic(err)
	}
	defer romFile.Close()
	for i := 0; i < len(data); i++ {
		memory.Poke(langRomStart+uint16(i), data[i])
	}

	/*
		Entry points as defined on the "BBC Microcomputer System
		User Guide", page 444

		Some vectors do not have an entrypoint associated we will
		assign a free one:
			UPTV - > OXUPT  at 0xff00
			EVNTV -> OXEVNT at 0xff03
			FSCV  -> OXFSC  at 0xff06
			IRQ2V -> OXIRQ2 at 0xff09
			IRQ1V -> IXIRQ1 at 0xff0c
			BRKV  -> OXBRK  at 0xff0f ;Break from the MOS break vector
			USERV -> OXUSER as 0xff12
			break -> OXBRKH at 0xff15 ;Break from the 6502 break vector
	*/

	// Setup MOS entry points, just having RTS is enough
	for i := mosVectors; i < vectorReset; i++ {
		memory.Poke(i, 0x60 /*RTS*/)
	}
	//memory.Poke(0xff0f, 0x40 /*RTI*/)

	// Setup the vectors in page 2 to the addresses on 0xffxx
	pokeWord(memory, 0x0222, 0xff00) // OXUPT  vector
	pokeWord(memory, 0x0220, 0xff03) // OXEVNT vector
	pokeWord(memory, 0x021E, 0xff06) // OXFSC  vector
	pokeWord(memory, 0x021C, 0xffce) // OSFIND vector
	pokeWord(memory, 0x021A, 0xffd1) // OSGBPB vector
	pokeWord(memory, 0x0218, 0xffd4) // OSBPUT vector
	pokeWord(memory, 0x0216, 0xffd7) // OSBGET vector
	pokeWord(memory, 0x0214, 0xffda) // OSARGS vector
	pokeWord(memory, 0x0212, 0xffdd) // OSFILE vector
	pokeWord(memory, 0x0210, 0xffe0) // OSRDCH vector
	pokeWord(memory, 0x020E, 0xffee) // OSWRCH vector
	pokeWord(memory, 0x020C, 0xfff1) // OSWORD vector
	pokeWord(memory, 0x020A, 0xfff4) // OSBYTE vector
	pokeWord(memory, 0x0208, 0xfff7) // OSCLI  vector
	pokeWord(memory, 0x0206, 0xff09) // OXIRQ2 vector
	pokeWord(memory, 0x0204, 0xff0c) // OXIRQ1 vector
	pokeWord(memory, 0x0202, 0xff0f) // OXBRK  vector
	pokeWord(memory, 0x0200, 0xff12) // OXUSER vector

	// Break vector
	pokeWord(memory, vectorBreak, 0x0ff15)

	// 6502 reset vector to point to the loaded ROM entry point
	memory.Poke(vectorReset, uint8(langRomStart&0xff))
	memory.Poke(vectorReset+1, uint8(langRomStart>>8))
	cpu.Reset()
	cpu.SetAXYP(1, 0, 0, 0) // ROM wants a 0x01 in A

	// Execute
	for {
		cpu.ExecuteInstruction()

		pc, sp := cpu.GetPCAndSP()
		if pc >= mosVectors {
			//cpu.SetTrace(false)
			a, x, y, p := cpu.GetAXYP()
			log := ""
			isIO := false

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
				*/
				address := peekWord(memory, 0x100+uint16(sp+2)) - 2 // BRK pushes the address then P
				faultNumber := memory.Peek(address + 1)
				faultMessage := address + 2
				faultString := getStringFromMem(memory, faultMessage, 0)

				pokeWord(memory, zpErrorPointer, address)
				brkv := peekWord(memory, mosVectorBrk)
				cpu.SetPC(brkv)
				log = fmt.Sprintf("BREAK(ERR=%02x, '%s'", faultNumber, faultString)

			case 0xffda: // OSARGS
				/*
					Call address &FFDA Indirected through &214
					This routine reads or writes an open file's attributes.
					On entry, X points to a four byte zero page control block.
					Y contains the file handle as provided by OSFIND, or zero.
					The accumulator contains a number specifying the action required.
				*/
				if y == 0 {
					switch a {
					case 0:
						// Returns the current filing system in A
						filingSystem := 4 // Disc filling system
						log = fmt.Sprintf("OSARGS('Get fiing system',A=%02x,Y=%02x)= %v", a, y, filingSystem)
					case 0xff:
						/*
							Update all files onto the media, ie ensure that the latest copy
							of the memory buffer is saved.
							We will do nothing.
						*/
						log = "OSARGS('Update all files onto the media')"
					default:
						panic(fmt.Sprintf("OSARGS(A=%02x,Y=%02x) not implemented", a, y))
					}
				} else {
					panic(fmt.Sprintf("OSARGS(A=%02x,Y=%02x) not implemented", a, y))
				}

			case 0xffdd: // OSFILE: Load or save a complete file. BPUG page 446
				/*
					http://beebwiki.mdfs.net/OSFILE
				*/
				controlBlock := uint16(x) + uint16(y)<<8
				filenameAddress := peekWord(memory, controlBlock)
				loadAddress := peekWord(memory, controlBlock+0x2)
				executionAddress := peekWord(memory, controlBlock+0x6)
				startAddress := peekWord(memory, controlBlock+0xa)
				endAddress := peekWord(memory, controlBlock+0xe)

				filename := getStringFromMem(memory, filenameAddress, 0x0d)
				filesize := endAddress - startAddress

				switch a {
				case 0:
					/*
						Save a block of memory as a file using the
						information provided in the parameter block.
					*/
					data := getMemSlice(memory, startAddress, filesize)
					err := ioutil.WriteFile(filename, data, 0644)
					if err != nil {
						panic(err)
					}
				case 0xff: // Load file into memory
					/*
						Load the named file, the address to which the file is
						loaded being determined by the lowest byte of the
						execution address in the control block (XY+6). If
						this byte is zero, the address given in the control
						block is used, otherwise the file's own load address
						is used.
					*/
					useLoadAddress := (executionAddress & 0xff) == 0
					if !useLoadAddress {
						panic("Loading files on their own load address not simplemented")
					}
					data, err := ioutil.ReadFile(filename)
					if err != nil {
						panic(err)
					}
					// NOTE: There is no maxLength?
					storeSliceinMem(memory, loadAddress, uint16(len(data)), data)
					filesize = uint16(len(data))
				}

				pokeWord(memory, controlBlock+0xa, filesize)
				pokeWord(memory, controlBlock+0x3, 0x33 /*attributes*/)

				log = fmt.Sprintf("OSFILE(A=%02x,FCB=%04x,FILE=%s,SIZE=%v)", a, controlBlock, filename, filesize)

			case 0xffe0: // OSRDCH
				/*
					This routine reads a character from the currently selected input
					stream and returns the character read in the accumulator.
					After an OSRDCH call: C=0 indicates that a valid character has
					been read; C=1 flags an error condition, A contains an error number.
				*/
				console.Scan()
				line := console.Text()
				// TODO: capture keystrokes. We will just get the first chat of the line
				// and ignore the rest.
				ch := strings.ToUpper(line)[0]
				pOut := p &^ 1 // Clear carry
				cpu.SetAXYP(ch, x, y, pOut)

				log = fmt.Sprintf("OSRDCH()=0x%02x", ch)

			case 0xffe3: // OSASCI
				/*
					This routine performs an OSWRCH call with the accumulator
					contents unless called with accumulator contents of &0D (13)
					when an OSNEWL call is performed.
				*/
				ch := string(a)
				if a == 0x0d {
					fmt.Println()
				} else {
					fmt.Printf("%v", ch)
				}

				if !unicode.IsGraphic([]rune(ch)[0]) {
					ch = ""
				}
				log = fmt.Sprintf("OSWRCH(0x%02x, '%v')", a, string(a))
				isIO = true

			case 0xffe7: // OSNEWL
				/*
					This call issues an LF CR (line feed, carriage return) to the currently
					selected output stream.
				*/
				fmt.Println()

				log = "OSNEWL()"
				isIO = true

			case 0xffee: // OSWRCH
				/*
					This call writes the character in A to the currently selected output
					stream.
					On exit A, X and Y are preserved, C, N, V and Z are undefined and D=0.
				*/
				ch := string(a)
				fmt.Printf("%v", ch)

				if !unicode.IsGraphic([]rune(ch)[0]) {
					ch = ""
				}
				log = fmt.Sprintf("OSWRCH(0x%02x, '%v')", a, string(a))
				isIO = true

			case 0xfff1: //OSWORD
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
					console.Scan()
					line := console.Text()
					line = strings.ToUpper(line)

					// TODO: check max size
					buffer := peekWord(memory, xy)
					maxLength := memory.Peek(xy + 2)
					minChar := memory.Peek(xy + 3)
					maxChar := memory.Peek(xy + 4)
					putStringInMem(memory, buffer, line+"\r", maxLength)
					pOut := p &^ 1 // Clear carry
					cpu.SetAXYP(1, x, uint8(len(line)), pOut)

					log = fmt.Sprintf("OSWORD00('read line',BUF=0x%04x,range=%02x-%02x, maxlen=%v)='%s'",
						buffer, minChar, maxChar, maxLength, line)

				case 0x01: // Read system clock
					/*
						This routine may be used to read the system clock (used for the
						TIME function in BASIC). The five byte clock value is written to
						the address contained in the X and Y registers. This clock is
						incremented every hundredth of a second and is set to 0 by a
						hard BREAK.
					*/
					duration := time.Since(referenceTime)
					ticks := duration.Milliseconds() / 10
					ticksLog := ticks & 0xff_ffff_ffff // 5 bytes
					buffer := peekWord(memory, xy)
					for i := uint16(0); i < 5; i++ {
						memory.Poke(buffer+i, uint8(ticks&0xff))
						ticks >>= 8
					}

					log = fmt.Sprintf("OSWORD01('read system clock',BUF=0x%04x)=%v", buffer, ticksLog)

				case 0x02: // Write system clock
					/*
						This routine may be used to set the system clock to a five byte
						value contained in memory at the address contained in the X
						and Y registers.
					*/
					buffer := peekWord(memory, xy)
					ticks := uint64(0)
					for i := uint16(0); i < 5; i++ {
						ticks <<= 8
						ticks += uint64(memory.Peek(buffer + i))
					}
					duration := time.Duration(ticks * 100 * 1000)
					referenceTime = time.Now()
					referenceTime.Add(duration * -1)

					log = fmt.Sprintf("OSWORD02('write system clock',TICKS=%v)", ticks)

				case 0x05: // Read I/O processor memory
					/*
						A byte of I/O processor memory may be read across the Tube
						using this call. A 32 bit address should be contained in memory
						at the address contained in the X and Y registers.
						On exit, The byte read will be contained in location XY+4.
					*/
					address := uint32(peekWord(memory, xy)) +
						uint32(peekWord(memory, xy+2))<<16
					value := memory.Peek((uint16(address)))
					memory.Poke(xy+4, value)

					log = fmt.Sprintf("OSWORD05('Read I/O processor memory',ADDRESS=0x%08x)=0x%02x",
						address, value)

				default:
					panic(fmt.Sprintf("OSWORD%02x call not supported", a))
				}

			case 0xfff4: // OSBYTE, page 408 BPUG
				option := "unknonw"
				switch a {
				case 0x7e:
					option = "Ack detection of an ESC condition"
					/*
						If an ESCAPE condition is detected, all active buffers will be flushed
						and any open EXEC files will be closed. There are no entry conditions;
						on exit, X=255 if the ESCAPE condition existed (X=0 otherwise), A is
						preserved, Y and C are undefined
					*/
					cpu.SetAXYP(a, 0, y, p)
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
					cpu.SetAXYP(a, 0xff, 0xff, p)
				case 0x83:
					option = "Read bottom of user mem"
					cpu.SetAXYP(a,
						uint8(userMemBottom&0xff),
						uint8(userMemBottom>>8),
						p)
				case 0x84:
					option = "Read top of user mem"
					cpu.SetAXYP(a,
						uint8(langRomStart&0xff),
						uint8(langRomStart>>8),
						p)
				case 0x85:
					option = "Read top of user mem for mode"
					cpu.SetAXYP(a,
						uint8(langRomStart&0xff),
						uint8(langRomStart>>8),
						p)
				case 0xda:
					option = "R/W number of items in VDU"
					/*
						Writing 0 to this location can be a useful way of abandoning a
						VDU queue otherwise writing to this location is not
						recommended.
					*/
					if x != 0 || y != 0 {
						panic("OSBYTEda only supported for x=0 and y=0")
					}
					// TODO; complete
				default:
					panic(fmt.Sprintf("OSBYTE%02x call not supported", a))
				}
				log = fmt.Sprintf("OSBYTE%02x('%s',X=0x%02x,Y=0x%02x)", a, option, x, y)

			case 0xfff7:
				/*
					This routine passes a line of text to the command line
					interpreter which decodes and executes any command
					recognised.
					X and Y should point to a line of text terminated by a
					carriage return character (ASCII &0D/13)
				*/
				xy := uint16(x) + uint16(y)<<8
				command := getStringFromMem(memory, xy, 0x0d)
				switch command {
				case "*HELP":
					fmt.Printf("bbz - Acorn MOS for 6502 adaptation layer, https://github.com/ivanizag/bbz\n")
				}
				log = fmt.Sprintf("OSCLI('%s')", command)

			default:
				panic(fmt.Sprintf("MOS call to 0x%04x not implemented", pc))
			}
			if apiLog && (!isIO || apiLogIO) {
				fmt.Printf("[[[%s]]]\n", log)
			}
		}
	}
}

func putStringInMem(mem core6502.Memory, address uint16, s string, maxLength uint8) {
	for i := 0; i < len(s) && i < int(maxLength); i++ {
		mem.Poke(address+uint16(i), s[i])
	}
}

func getStringFromMem(mem core6502.Memory, address uint16, terminator uint8) string {
	str := ""
	for {
		ch := mem.Peek(address)
		//fmt.Printf("{%04x: %02x\n", address, ch)
		if ch == terminator {
			break
		}
		str += string(ch)
		address++
	}
	return str
}

func getMemSlice(mem core6502.Memory, address uint16, length uint16) []uint8 {
	slice := make([]uint8, 0, length)
	for i := uint16(0); i < length; i++ {
		slice = append(slice, mem.Peek(address+i))
	}
	return slice
}

func storeSliceinMem(mem core6502.Memory, address uint16, maxLength uint16, data []uint8) uint16 {
	var i uint16
	for i = 0; i < maxLength && i < uint16(len(data)); i++ {
		mem.Poke(address+i, data[i])
	}
	return i
}

func peekWord(mem core6502.Memory, address uint16) uint16 {
	return uint16(mem.Peek(address)) + uint16(mem.Peek(address+1))<<8
}

func pokeWord(mem core6502.Memory, address uint16, value uint16) {
	mem.Poke(address, uint8(value&0xff))
	mem.Poke(address+1, uint8(value>>8))
}
