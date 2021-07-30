package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/ivanizag/izapple2/core6502"
)

const (
	zpAccumulator  uint16 = 0x00fc
	zpErrorPointer uint16 = 0x00fd
	mosVectorBrk   uint16 = 0x0202

	userMemBottom       uint16 = 0x0e00
	langRomStart        uint16 = 0x8000
	langCopyrightOffset uint16 = 0x8007
	langRomName         uint16 = 0x8009

	mosVectors      uint16 = 0xff00
	breakEntryPoint uint16 = 0xffb0 // Invented as there is not an address defined
	vectorReset     uint16 = 0xfffc
	vectorBreak     uint16 = 0xfffe
)

type environment struct {
	cpu           *core6502.State
	mem           core6502.Memory
	vdu           *vdu
	in            *bufio.Scanner
	referenceTime time.Time
	apiLog        bool
	apiLogIO      bool
	panicOnErr    bool
}

func (env *environment) loadRom(filename string) {
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
		env.mem.Poke(langRomStart+uint16(i), data[i])
	}

	// See http://www.sprow.co.uk/bbc/library/sidewrom.pdf
	language := env.getStringFromMem(langRomName, 0)
	copyrightAddress := 0x8000 + 1 + uint16(env.mem.Peek(langCopyrightOffset))
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

func (env *environment) loadMos() {
	// Setup MOS entry points, just having RTS is enough
	for i := mosVectors; i < vectorReset; i++ {
		env.mem.Poke(i, 0x60 /*RTS*/)
	}

	// Setup the vectors in page 2 to the addresses on 0xffxx
	env.pokeWord(0x0222, 0xff00) // OXUPT  vector
	env.pokeWord(0x0220, 0xff03) // OXEVNT vector
	env.pokeWord(0x021E, 0xff06) // OXFSC  vector
	env.pokeWord(0x021C, 0xffce) // OSFIND vector
	env.pokeWord(0x021A, 0xffd1) // OSGBPB vector
	env.pokeWord(0x0218, 0xffd4) // OSBPUT vector
	env.pokeWord(0x0216, 0xffd7) // OSBGET vector
	env.pokeWord(0x0214, 0xffda) // OSARGS vector
	env.pokeWord(0x0212, 0xffdd) // OSFILE vector
	env.pokeWord(0x0210, 0xffe0) // OSRDCH vector
	env.pokeWord(0x020E, 0xffee) // OSWRCH vector
	env.pokeWord(0x020C, 0xfff1) // OSWORD vector
	env.pokeWord(0x020A, 0xfff4) // OSBYTE vector
	env.pokeWord(0x0208, 0xfff7) // OSCLI  vector
	env.pokeWord(0x0206, 0xff09) // OXIRQ2 vector
	env.pokeWord(0x0204, 0xff0c) // OXIRQ1 vector
	env.pokeWord(0x0202, 0xff0f) // OXBRK  vector
	env.pokeWord(0x0200, 0xff12) // OXUSER vector

	// Break vector
	env.pokeWord(vectorBreak, 0x0ff15)

	// 6502 reset vector to point to the loaded ROM entry point
	env.mem.Poke(vectorReset, uint8(langRomStart&0xff))
	env.mem.Poke(vectorReset+1, uint8(langRomStart>>8))
	env.cpu.Reset() // It has to be done after the reset vector is set.

	// ROM wants a 0x01 in A, see http://beebwiki.mdfs.net/Paged_ROM
	env.cpu.SetAXYP(1, 0, 0, 0)

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
	fmt.Printf("[[[%s]]]\n", msg)
}

func (env *environment) putStringInMem(address uint16, s string, maxLength uint8) {
	for i := 0; i < len(s) && i < int(maxLength); i++ {
		env.mem.Poke(address+uint16(i), s[i])
	}
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
