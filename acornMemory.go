package main

import "io/ioutil"

// FlatMemory puts RAM on the 64Kb addressable by the processor
type AcornMemory struct {
	data [65536]uint8
}

// Peek returns the data on the given address
func (m *AcornMemory) Peek(address uint16) uint8 {
	return m.data[address]
}

// PeekCode returns the data on the given address
func (m *AcornMemory) PeekCode(address uint16) uint8 {
	return m.data[address]
}

// Poke sets the data at the given address
func (m *AcornMemory) Poke(address uint16, value uint8) {
	m.data[address] = value
}

func (m *AcornMemory) loadBinary(filename string) error {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	for i, v := range bytes {
		m.Poke(uint16(i), uint8(v))
	}

	return nil
}

func (m *AcornMemory) getString(address uint16, terminator uint8) string {
	str := ""
	for {
		ch := m.Peek(address)
		//fmt.Printf("{%04x: %02x\n", address, ch)
		if ch == terminator {
			break
		}
		str += string(ch)
		address++
	}
	return str
}

func (m *AcornMemory) storeString(address uint16, s string, terminator uint8, maxLength uint8) {
	// maxLength not including terminator
	var i int
	for i = 0; i < len(s) && i < int(maxLength); i++ {
		m.Poke(address+uint16(i), s[i])
	}
	m.Poke(address+uint16(i), terminator)
}

func (m *AcornMemory) getSlice(address uint16, length uint16) []uint8 {
	slice := make([]uint8, 0, length)
	for i := uint16(0); i < length; i++ {
		slice = append(slice, m.Peek(address+i))
	}
	return slice
}

func (m *AcornMemory) storeSlice(address uint16, maxLength uint16, data []uint8) uint16 {
	var i uint16
	for i = 0; i < maxLength && i < uint16(len(data)); i++ {
		m.Poke(address+i, data[i])
	}
	return i
}

func (m *AcornMemory) peekWord(address uint16) uint16 {
	return uint16(m.Peek(address)) + uint16(m.Peek(address+1))<<8
}

func (m *AcornMemory) pokeWord(address uint16, value uint16) {
	m.Poke(address, uint8(value&0xff))
	m.Poke(address+1, uint8(value>>8))
}

func (m *AcornMemory) peeknbytes(address uint16, n int) uint64 {
	ticks := uint64(0)
	for i := n - 1; i >= 0; i-- {
		ticks <<= 8
		ticks += uint64(m.Peek(address + uint16(i)))
	}
	return ticks
}

func (m *AcornMemory) pokenbytes(address uint16, n int, value uint64) {
	for i := 0; i < n; i++ {
		m.Poke(address+uint16(i), uint8(value&0xff))
		value >>= 8
	}
}
