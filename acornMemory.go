package main

import (
	_ "embed"
	"fmt"
	"os"
)

type acornMemory struct {
	data            [65536]uint8
	sideRom         [16][]uint8
	writeProtectRom [16]bool
	activeRom       uint8
	memLog          bool

	// Some (probably unneeded) optimisations
	pActiveRom   *[]uint8
	activeRomEnd uint16
}

func newAcornMemory(memLog bool) *acornMemory {
	var a acornMemory
	a.memLog = memLog
	a.selectRom(0xf)
	return &a
}

func (m *acornMemory) Poke(address uint16, value uint8) {
	if m.memLog {
		area := memoryArea(address)
		if area != "" {
			fmt.Printf("[[[Poke(%s:%02x, %02x]]]\n", area, address&0xff, value)
		}
	}
	if address >= romStartAddress {
		if address <= m.activeRomEnd {
			if !m.writeProtectRom[m.activeRom] {
				(*m.pActiveRom)[address-romStartAddress] = value
			}
			return
		}

		if address == sheilaRomLatch {
			m.selectRom(value & 0xf)
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

	if romStartAddress <= address && address <= m.activeRomEnd {
		return (*m.pActiveRom)[address-romStartAddress]
	}

	return value
}

func (m *acornMemory) PeekCode(address uint16) uint8 {
	return m.Peek(address)
}

func (m *acornMemory) selectRom(slot uint8) {
	m.activeRom = slot
	if len(m.sideRom[m.activeRom]) > 0 {
		m.pActiveRom = &m.sideRom[slot]
		m.activeRomEnd = romStartAddress + uint16(len(*m.pActiveRom)) - 1
	} else {
		m.pActiveRom = nil
		m.activeRomEnd = romStartAddress - 1
	}

	if m.memLog {
		fmt.Printf("[[[EnableROM(%x)]]]\n", m.activeRom)
	}
}

func memoryArea(address uint16) string {
	if address == 0x00ff {
		return "ESCAPE"
	}
	if address >= mosVariablesStart && address <= mosVariablesEnd {
		return "MOSVARS"
	}

	switch address >> 8 {
	//case 0x00:
	//	return "ZEROPAGE"
	case 0xfc:
		return "FRED"
	case 0xfd:
		return "JIM"
	case 0xfe:
		return "SHEILA"
	}
	return ""
}

//go:embed firmware
var firmware []byte

func (m *acornMemory) loadFirmware() {
	//for i := 0; i < len(firmware); i++ {
	//	m.data[i] = firmware[i]
	//}
	copy(m.data[:], firmware)
}

func (m *acornMemory) loadRom(filename string, slot uint8) {
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("ROM can't be loaded:\n    %s\n", err)
		os.Exit(1)
	}

	m.sideRom[slot] = data
	m.writeProtectRom[slot] = true

	// Cache the ROM type
	m.data[mosRomTypeTable+uint16(slot)] = data[romTypeByte-romStartAddress]
}

func (m *acornMemory) completeWithRam() {
	for i := 0; i < 16; i++ {
		slot := m.sideRom[i]
		if len(slot) == 0 {
			m.sideRom[i] = make([]uint8, 16*1024)
		}
	}
}

// Memory access helpers, peek and pokes

func (m *acornMemory) peekString(address uint16, terminator uint8) string {
	str := ""
	for {
		ch := m.Peek(address) & 0x7f
		if ch == terminator {
			break
		}
		str += string(ch)
		address++
	}
	return str
}

func (m *acornMemory) pokeString(address uint16, s string, terminator uint8, maxLength uint8) {
	// maxLength not including terminator
	var i int
	for i = 0; i < len(s) && i < int(maxLength); i++ {
		m.Poke(address+uint16(i), s[i])
	}
	m.Poke(address+uint16(i), terminator)
}

func (m *acornMemory) peekSlice(address uint16, length uint16) []uint8 {
	slice := make([]uint8, 0, length)
	for i := uint16(0); i < length; i++ {
		slice = append(slice, m.Peek(address+i))
	}
	return slice
}

func (m *acornMemory) pokeSlice(address uint16, maxLength uint16, data []uint8) uint16 {
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

func (m *acornMemory) peekDoubleWord(address uint16) uint32 {
	return uint32(m.Peek(address)) +
		uint32(m.Peek(address+1))<<8 +
		uint32(m.Peek(address+2))<<16 +
		uint32(m.Peek(address+3))<<24
}

func (m *acornMemory) pokeDoubleWord(address uint16, value uint32) {
	m.Poke(address, uint8(value))
	m.Poke(address+1, uint8(value>>8))
	m.Poke(address+2, uint8(value>>16))
	m.Poke(address+3, uint8(value>>24))
}

func (m *acornMemory) peekNBytes(address uint16, n int) uint64 {
	ticks := uint64(0)
	for i := n - 1; i >= 0; i-- {
		ticks <<= 8
		ticks += uint64(m.Peek(address + uint16(i)))
	}
	return ticks
}

func (m *acornMemory) pokeNBytes(address uint16, n int, value uint64) {
	for i := 0; i < n; i++ {
		m.Poke(address+uint16(i), uint8(value&0xff))
		value >>= 8
	}
}

func (m *acornMemory) pokeBCD(address uint16, value uint8) {
	bcd := value/10*16 + value%10
	m.Poke(address, bcd)
}
