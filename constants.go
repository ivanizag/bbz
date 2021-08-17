package main

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
	zpStr          uint16 = 0x00f2 // OSCLI command line
	zpROMSelect    uint16 = 0x00f4
	zpAccumulator  uint16 = 0x00fc
	zpErrorPointer uint16 = 0x00fd

	vectorBRK    uint16 = 0x0202
	romTypeTable uint16 = 0x023a

	// ROM header https://tobylobster.github.io/mos/mos/S-s2.html#SP26
	userMemBottom             uint16 = 0x0e00
	romStartAddress           uint16 = 0x8000
	romEndAddress             uint16 = 0xbfff
	romServiceEntry           uint16 = 0x8003
	romTypeByte               uint16 = 0x8006
	romCopyrightOffsetPointer uint16 = 0x8007
	romTitleString            uint16 = 0x8009

	// Support code on the firmware. Check firmware.lst when changing firmware.s
	procCLITOROMS uint16 = 0xf000

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

	// Fred, Jim and Sheila
	sheila_rom_latch uint16 = 0xfe30

	maxFiles        = 5
	errorTodo uint8 = 129 // TODO: find proper error number
)
