package main

import (
	"fmt"
	"io/ioutil"
)

/*
Memory map, pages:
	00: zero page
	01: stack
	02: mos vectors and other
	...
	80: loaded rom
	...
	fa: brk error messages
	fb: entrypoints trapped by the host
	fc: FRED
	fd: JIM
	fe: SHEILA
	ff: mos and 6502 entrypoints

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

	// ROM header https://tobylobster.github.io/mos/mos/S-s2.html#SP26
	userMemBottom             uint16 = 0x0e00
	romStartAddress           uint16 = 0x8000
	romServiceEntry           uint16 = 0x8003
	romTypeByte               uint16 = 0x8006
	romCopyrightOffsetPointer uint16 = 0x8007
	romTitleString            uint16 = 0x8009

	// Scratch area for errors in page 0xfa
	errorArea             uint16 = 0xfa00
	errorMessageMaxLength int    = 100

	// Entry points for host interception in page 0xfb
	entryPoints       uint16 = 0xfb00
	epUPT             uint16 = 0xfb00
	epEVNT            uint16 = 0xfb01
	epFSC             uint16 = 0xfb02
	epFIND            uint16 = 0xfb03
	epGBPB            uint16 = 0xfb04
	epBPUT            uint16 = 0xfb05
	epBGET            uint16 = 0xfb06
	epARGS            uint16 = 0xfb07
	epFILE            uint16 = 0xfb08
	epRDCH            uint16 = 0xfb09
	epWRCH            uint16 = 0xfb0a
	epWORD            uint16 = 0xfb0b
	epBYTE            uint16 = 0xfb0c
	epCLI             uint16 = 0xfb0d
	epIRQ2            uint16 = 0xfb0e
	epIRQ1            uint16 = 0xfb0f
	epBRK             uint16 = 0xfb10
	epUSER            uint16 = 0xfb11
	epSYSBRK          uint16 = 0xfb12
	epRDRM            uint16 = 0xfb13
	epVDUCH           uint16 = 0xfb14
	epGSINIT          uint16 = 0xfb16
	epGSREAD          uint16 = 0xfb17
	epNET             uint16 = 0xfb18
	epVDU             uint16 = 0xfb19
	epKEY             uint16 = 0xfb1a
	epINS             uint16 = 0xfb1b
	epREM             uint16 = 0xfb1c
	epCNP             uint16 = 0xfb1d
	epIND1            uint16 = 0xfb1e
	epIND2            uint16 = 0xfb1f
	epIND3            uint16 = 0xfb20
	epEntryPointsLast uint16 = 0xfb20

	// MOS entrypoints and 6502 vectors
	mosVectors  uint16 = 0xff00
	vectorReset uint16 = 0xfffc
	vectorBreak uint16 = 0xfffe

	maxFiles        = 5
	errorTodo uint8 = 129 // TODO: find proper error number
)

func loadMosFromFile(env *environment, firmFilename string) {
	data, err := ioutil.ReadFile(firmFilename)
	if err != nil {
		panic(err)
	}

	for i := 0; i < len(data); i++ {
		env.mem.Poke(uint16(i), data[i])
	}
}

func loadRom(env *environment, filename string) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

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
