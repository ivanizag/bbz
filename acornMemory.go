package main

import (
	"fmt"
	"io/ioutil"
)

type acornMemory struct {
	data      [65536]uint8
	sideRom   [16][]uint8
	activeRom uint8
	memLog    bool
}

func newAcornMemory(memLog bool) *acornMemory {
	var a acornMemory
	a.memLog = memLog
	a.activeRom = 0xf
	return &a
}

func (m *acornMemory) Poke(address uint16, value uint8) {
	if m.memLog {
		area := memoryArea(address)
		if area != "" {
			fmt.Printf("[[[Poke(%s:%02x, %02x]]]\n", area, address&0xff, value)
		}
	}

	if romStartAddress <= address && address <= romEndAddress {
		m.sideRom[m.activeRom][address-romStartAddress] = value
		return
	}

	if address == sheila_rom_latch {
		m.activeRom = value & 0xf
		if m.memLog {
			fmt.Printf("[[[EnableROM(%x)]]]\n", m.activeRom)
		}
	}
	m.data[address] = value
}

func (m *acornMemory) Peek(address uint16) uint8 {
	value := m.data[address]

	if m.memLog {
		area := memoryArea(address)
		if area != "" {
			fmt.Printf("[[[Peek(%s:%02x) => %02x]]]\n", area, address&0xff, value)
		}
	}

	if romStartAddress <= address && address <= romEndAddress {
		return m.sideRom[m.activeRom][address-romStartAddress]
	}

	return value
}

func (m *acornMemory) PeekCode(address uint16) uint8 {
	return m.Peek(address)
}

func memoryArea(address uint16) string {
	switch address >> 8 {
	case 0xfc:
		return "FRED"
	case 0xfd:
		return "JIM"
	case 0xfe:
		return "SHEILA"
	}
	return ""
}

func (m *acornMemory) loadFirmware(firmFilename string) {
	data, err := ioutil.ReadFile(firmFilename)
	if err != nil {
		panic(err)
	}

	for i := 0; i < len(data); i++ {
		m.data[i] = data[i]
	}
}

func (m *acornMemory) loadRom(filename string, slot uint8) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	m.sideRom[slot] = data

	// Cache the ROM type
	m.data[romTypeTable+uint16(slot)] = data[romTypeByte-romStartAddress]
}

// helpers

func (m *acornMemory) getString(address uint16, terminator uint8) string {
	str := ""
	for {
		ch := m.Peek(address) & 0x7f
		//fmt.Printf("{%04x: %02x\n", address, ch)
		if ch == terminator {
			break
		}
		str += string(ch)
		address++
	}
	return str
}

func (m *acornMemory) storeString(address uint16, s string, terminator uint8, maxLength uint8) {
	// maxLength not including terminator
	var i int
	for i = 0; i < len(s) && i < int(maxLength); i++ {
		m.Poke(address+uint16(i), s[i])
	}
	m.Poke(address+uint16(i), terminator)
}

func (m *acornMemory) getSlice(address uint16, length uint16) []uint8 {
	slice := make([]uint8, 0, length)
	for i := uint16(0); i < length; i++ {
		slice = append(slice, m.Peek(address+i))
	}
	return slice
}

func (m *acornMemory) storeSlice(address uint16, maxLength uint16, data []uint8) uint16 {
	var i uint16
	for i = 0; i < maxLength && i < uint16(len(data)); i++ {
		m.Poke(address+i, data[i])
	}
	return i
}

func (m *acornMemory) peekWord(address uint16) uint16 {
	return uint16(m.Peek(address)) + uint16(m.Peek(address+1))<<8
}

func (m *acornMemory) pokeWord(address uint16, value uint16) {
	m.Poke(address, uint8(value&0xff))
	m.Poke(address+1, uint8(value>>8))
}

func (m *acornMemory) peeknbytes(address uint16, n int) uint64 {
	ticks := uint64(0)
	for i := n - 1; i >= 0; i-- {
		ticks <<= 8
		ticks += uint64(m.Peek(address + uint16(i)))
	}
	return ticks
}

func (m *acornMemory) pokenbytes(address uint16, n int, value uint64) {
	for i := 0; i < n; i++ {
		m.Poke(address+uint16(i), uint8(value&0xff))
		value >>= 8
	}
}
