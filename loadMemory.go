package main

import (
	"fmt"
	"io/ioutil"
	"os"
)

/*
Memory map:
	0000-00ff: zero page
	0100-01ff: stack
	0200-02ff: mos vectors and other
	...
	8000-bfff: loaded rom
	...
	fd00-fdff: brk error messages
	fe00-feff: entrypoints trapped by the host
	ff00-ffff: mos and 6502 entrypoints

*/
const (
	zpAccumulator  uint16 = 0x00fc
	zpErrorPointer uint16 = 0x00fd

	// Vectors in page 2 that can be changed by the program being run
	vectorUSERV uint16 = 0x0200 // User vector
	vectorBRKV  uint16 = 0x0202 // BRK vector
	vectorIRQ1V uint16 = 0x0204 // Primary IRQ vector
	vectorIRQ2V uint16 = 0x0206 // Unrecognised IRQ vector
	vectorCLIV  uint16 = 0x0208 // Command line interpreter
	vectorBYTEV uint16 = 0x020A // OSBYTE call
	vectorWORDV uint16 = 0x020C // OSWORD call
	vectorWRCHV uint16 = 0x020E // OSWRCH call
	vectorRDCHV uint16 = 0x0210 // OSRDCH call
	vectorFILEV uint16 = 0x0212 // Load / Save file
	vectorARGSV uint16 = 0x0214 // Load / Save file parameters
	vectorBGETV uint16 = 0x0216 // Get byte from file
	vectorBPUTV uint16 = 0x0218 // Put byte to file
	vectorGBPBV uint16 = 0x021A // Transfer data to or from a file
	vectorFINDV uint16 = 0x021C // Open / Close file
	vectorFSCV  uint16 = 0x021E // Filing system control
	vectorEVNTV uint16 = 0x0220 // Events
	vectorUPTV  uint16 = 0x0222 // User print
	vectorNETV  uint16 = 0x0224 // Econet
	vectorVDUV  uint16 = 0x0226 // Unrecognised PLOT / VDU 23 commands
	vectorKEYV  uint16 = 0x0228 // Keyboard
	vectorINSV  uint16 = 0x022A // Insert character into buffer
	vectorREMV  uint16 = 0x022C // Remove character from buffer
	vectorCNPV  uint16 = 0x022E // Count or purge buffer
	vectorIND1V uint16 = 0x0230 // Unused vector
	vectorIND2V uint16 = 0x0232 // Unused vector
	vectorIND3V uint16 = 0x0234 // Unused vector

	// ROM header
	userMemBottom             uint16 = 0x0e00
	romStartAddress           uint16 = 0x8000
	romCopyrightOffsetPointer uint16 = 0x8007
	romTitleString            uint16 = 0x8009

	// Scratch are for errors in page 0xfd
	errorArea             uint16 = 0xfd00
	errorMessageMaxLength int    = 100

	// Entry points for host interception in page 0xfe
	entryPoints uint16 = 0xfe00
	epUPT       uint16 = 0xfe00
	epEVNT      uint16 = 0xfe01
	epFSC       uint16 = 0xfe02
	epFIND      uint16 = 0xfe03
	epGBPB      uint16 = 0xfe04
	epBPUT      uint16 = 0xfe05
	epBGET      uint16 = 0xfe06
	epARGS      uint16 = 0xfe07
	epFILE      uint16 = 0xfe08
	epRDCH      uint16 = 0xfe09
	epWRCH      uint16 = 0xfe0a
	epWORD      uint16 = 0xfe0b
	epBYTE      uint16 = 0xfe0c
	epCLI       uint16 = 0xfe0d
	epIRQ2      uint16 = 0xfe0e
	epIRQ1      uint16 = 0xfe0f
	epBRK       uint16 = 0xfe10
	epUSER      uint16 = 0xfe11
	epSYSBRK    uint16 = 0xfe12
	epRDRM      uint16 = 0xfe13
	epVDUCH     uint16 = 0xfe14
	epGSINIT    uint16 = 0xfe16
	epGSREAD    uint16 = 0xfe17
	epNET       uint16 = 0xfe18
	epVDU       uint16 = 0xfe19
	epKEY       uint16 = 0xfe1a
	epINS       uint16 = 0xfe1b
	epREM       uint16 = 0xfe1c
	epCNP       uint16 = 0xfe1d
	epIND1      uint16 = 0xfe1e
	epIND2      uint16 = 0xfe1f
	epIND3      uint16 = 0xfe20

	// MOS entrypoints and 6502 vectors
	mosVectors  uint16 = 0xff00
	vectorReset uint16 = 0xfffc
	vectorBreak uint16 = 0xfffe

	maxFiles        = 5
	errorTodo uint8 = 129 // TODO: find proper error number
)

func loadRom(env *environment, filename string) {
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	data, err := ioutil.ReadAll(file)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	for i := 0; i < len(data); i++ {
		env.mem.Poke(romStartAddress+uint16(i), data[i])
	}

	// See http://www.sprow.co.uk/bbc/library/sidewrom.pdf
	language := env.getStringFromMem(romTitleString, 0)
	copyrightAddress := 0x8000 + 1 + uint16(env.mem.Peek(romCopyrightOffsetPointer))
	copyrigt := env.getStringFromMem(copyrightAddress, 0)

	/*
		Next, the MOS will set the error point at &FD/&FE to point at the version string (or copyright
		message if no version string is present).
	*/
	env.pokeWord(zpErrorPointer, copyrightAddress)
	/*
		The MOS also automatically prints the ROM's title string (&8009) so that the user is acknowledged.
	*/
	fmt.Printf("%s - %s\n", language, copyrigt)
}

func loadMos(env *environment) {

	// Setup the vectors in page 2 to the addresses on 0xffxx
	env.pokeWord(0x0222, epUPT)
	env.pokeWord(0x0220, epEVNT)
	env.pokeWord(0x021E, epFSC)
	env.pokeWord(0x021C, epFIND)
	env.pokeWord(0x021A, epGBPB)
	env.pokeWord(0x0218, epBPUT)
	env.pokeWord(0x0216, epBGET)
	env.pokeWord(0x0214, epARGS)
	env.pokeWord(0x0212, epFILE)
	env.pokeWord(0x0210, epRDCH)
	env.pokeWord(0x020E, epWRCH)
	env.pokeWord(0x020C, epWORD)
	env.pokeWord(0x020A, epBYTE)
	env.pokeWord(0x0208, epCLI)
	env.pokeWord(0x0206, epIRQ2)
	env.pokeWord(0x0204, epIRQ1)
	env.pokeWord(0x0202, epBRK)
	env.pokeWord(0x0200, epUSER)

	env.pokeWord(vectorUSERV, epUSER)
	env.pokeWord(vectorBRKV, epBRK)
	env.pokeWord(vectorIRQ1V, epIRQ1)
	env.pokeWord(vectorIRQ2V, epIRQ2)
	env.pokeWord(vectorCLIV, epCLI)
	env.pokeWord(vectorBYTEV, epBYTE)
	env.pokeWord(vectorWORDV, epWORD)
	env.pokeWord(vectorWRCHV, epWRCH)
	env.pokeWord(vectorRDCHV, epRDCH)
	env.pokeWord(vectorFILEV, epFILE)
	env.pokeWord(vectorARGSV, epARGS)
	env.pokeWord(vectorBGETV, epBGET)
	env.pokeWord(vectorBPUTV, epBPUT)
	env.pokeWord(vectorGBPBV, epGBPB)
	env.pokeWord(vectorFINDV, epFIND)
	env.pokeWord(vectorFSCV, epFSC)
	env.pokeWord(vectorEVNTV, epEVNT)
	env.pokeWord(vectorUPTV, epUPT)
	env.pokeWord(vectorNETV, epNET)
	env.pokeWord(vectorVDUV, epVDU)
	env.pokeWord(vectorKEYV, epKEY)
	env.pokeWord(vectorINSV, epINS)
	env.pokeWord(vectorREMV, epREM)
	env.pokeWord(vectorCNPV, epCNP)
	env.pokeWord(vectorIND1V, epIND1)
	env.pokeWord(vectorIND2V, epIND2)
	env.pokeWord(vectorIND3V, epIND3)

	// Setup host entry points in page 0xfe, just having RTS is enough to be trapped.
	for i := entryPoints; i < mosVectors; i++ {
		env.mem.Poke(i, 0x60 /*RTS*/)
	}

	// Setup 0xff page
	for i := mosVectors; i < vectorReset; i++ {
		// Prefill with nulls
		env.mem.Poke(i, 0x00 /*BRK*/)
	}

	// See: https://tobylobster.github.io/mos/mos/S-s24.html
	writeJMP(env, 0xffb9, epRDRM)
	writeJMP(env, 0xffbc, epVDUCH)
	writeJMP(env, 0xffbf, epEVNT)
	writeJMP(env, 0xffc2, epGSINIT)
	writeJMP(env, 0xffc5, epGSREAD)
	writeJMP(env, 0xffc8, epRDCH)
	writeJMP(env, 0xffcb, epWRCH)
	writeJMI(env, 0xffce, vectorFINDV)
	writeJMI(env, 0xffd1, vectorGBPBV)
	writeJMI(env, 0xffd4, vectorBPUTV)
	writeJMI(env, 0xffd7, vectorBGETV)
	writeJMI(env, 0xffda, vectorARGSV)
	writeJMI(env, 0xffdd, vectorFILEV)
	writeJMI(env, 0xffe0, vectorRDCHV)
	env.mem.Poke(0xffe3, 0xc9) // CMP #$0D
	env.mem.Poke(0xffe4, 0x0d)
	env.mem.Poke(0xffe5, 0xd0) // BNE $ffee
	env.mem.Poke(0xffe6, 0x07)
	env.mem.Poke(0xffe7, 0xa9) // LDA #$0A
	env.mem.Poke(0xffe8, 0x0a)
	env.mem.Poke(0xffe9, 0x20) // JSR $ffee
	env.mem.Poke(0xffea, 0xee)
	env.mem.Poke(0xffeb, 0xff)
	env.mem.Poke(0xffec, 0xa9) // LDA #$0D
	env.mem.Poke(0xffed, 0x0d)
	writeJMI(env, 0xffee, vectorWRCHV)
	writeJMI(env, 0xfff1, vectorWORDV)
	writeJMI(env, 0xfff4, vectorBYTEV)
	writeJMI(env, 0xfff7, vectorCLIV)

	env.pokeWord(vectorBreak, epSYSBRK)        // To trap breaks from the BRK opcode
	env.pokeWord(vectorReset, romStartAddress) // The reset vector to point to the loaded ROM entry point

	// ROM wants a 0x01 in A, see http://beebwiki.mdfs.net/Paged_ROM
	env.cpu.Reset()
	env.cpu.SetAXYP(1, 0, 0, 0)
}

func writeJMP(env *environment, address uint16, dest uint16) {
	env.mem.Poke(address, 0x4c /*JMP*/)
	env.pokeWord(address+1, dest)
}

func writeJMI(env *environment, address uint16, dest uint16) {
	env.mem.Poke(address, 0x6c /*JMP indirect*/)
	env.pokeWord(address+1, dest)
}
