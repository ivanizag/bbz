package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

const (
	osNotFound       uint8 = 0
	osFileFound      uint8 = 1
	osDirectoryFound uint8 = 2

	cbLoadAddress            uint16 = 0x2
	cbExecutionAddress       uint16 = 0x6
	cbStartAddressOrSize     uint16 = 0xa
	cbEndAddressOrAttributes uint16 = 0xe

	loadAddressNull      uint32 = ^uint32(0)
	executionAddressNull uint32 = ^uint32(0)
)

type fileAttributes struct {
	fileType         uint8
	fileSize         uint32
	hasMetadata      bool
	loadAddress      uint32
	executionAddress uint32
}

func execOSFILE(env *environment) {
	a, x, y, p := env.cpu.GetAXYP()

	// See: http://beebwiki.mdfs.net/OSFILE
	controlBlock := uint16(x) + uint16(y)<<8
	filenameAddress := env.mem.peekWord(controlBlock)
	loadAddress := uint32(env.mem.peeknbytes(controlBlock+cbLoadAddress, 4))
	executionAddress := uint32(env.mem.peeknbytes(controlBlock+cbExecutionAddress, 4))
	startAddress := uint32(env.mem.peeknbytes(controlBlock+cbStartAddressOrSize, 4))
	endAddress := uint32(env.mem.peeknbytes(controlBlock+cbEndAddressOrAttributes, 4))

	filename := env.mem.getString(filenameAddress, 0x0d)
	filesize := endAddress - startAddress

	newA := uint8(0) // Nothing found
	option := ""
	switch a {
	case 0:
		option = "Save file"
		/*
			Save a block of memory as a file using the
			information provided in the parameter block.
		*/
		data := env.mem.getSlice(uint16(startAddress), uint16(filesize))
		err := ioutil.WriteFile(filename, data, 0644)
		if err != nil {
			env.raiseError(errorTodo, err.Error())
			newA = osNotFound
		} else {
			newA = osFileFound
		}
	case 5:
		option = "File info"
		/*
			Read a file’s catalogue information, with the file
			type returned in the accumulator. The information
			is written to the parameter block.
		*/
		attr := getFileAttributes(env, filename)
		if attr.fileType != osNotFound {
			env.mem.pokenbytes(controlBlock+cbStartAddressOrSize, 4, uint64(attr.fileSize))
			if attr.hasMetadata {
				env.mem.pokenbytes(controlBlock+cbLoadAddress, 4, uint64(attr.loadAddress))
				env.mem.pokenbytes(controlBlock+cbExecutionAddress, 4, uint64(attr.executionAddress))
			}
		}
		newA = attr.fileType

	case 0xff: // Load file into memory
		option = "Load file"
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
			loadAddress = loadAddressNull
		}

		attr := loadFile(env, filename, loadAddress)
		env.mem.pokenbytes(controlBlock+cbStartAddressOrSize, 4, uint64(attr.fileSize))
		newA = attr.fileType

	default:
		env.notImplemented(fmt.Sprintf("OSFILE(A=%02x)", a))
	}

	env.mem.pokeWord(controlBlock+cbEndAddressOrAttributes, 0x00 /*attributes*/)

	env.cpu.SetAXYP(newA, x, y, p)
	env.log(fmt.Sprintf("OSFILE('%s',A=%02x,FCB=%04x,FILE=%s) => %v",
		option, a, controlBlock, filename, newA))

}

func loadFile(env *environment, filename string, loadAddress uint32) fileAttributes {
	attr := getFileAttributes(env, filename)
	if attr.fileType == osNotFound {
		return attr
	}
	if attr.fileType == osDirectoryFound {
		env.raiseError(errorTodo, "Directory found")
		return attr
	}

	if loadAddress == loadAddressNull {
		if !attr.hasMetadata {
			env.raiseError(errorTodo, "Missing metadata file")
			attr.fileType = osNotFound
			return attr
		}
		loadAddress = uint32(attr.loadAddress)
	}
	data, err := os.ReadFile(filename)
	if err != nil {
		env.raiseError(errorTodo, err.Error())
		attr.fileType = osNotFound
		return attr
	}

	// NOTE: There is no maxLength?
	env.mem.storeSlice(uint16(loadAddress), uint16(len(data)), data)
	return attr
}

func getFileAttributes(env *environment, filename string) fileAttributes {
	var attr fileAttributes

	fileInfo, err := os.Stat(filename)
	if errors.Is(err, os.ErrNotExist) {
		attr.fileType = osNotFound
		env.raiseError(214, "File not found")
		return attr
	}
	if err != nil {
		attr.fileType = osNotFound
		env.raiseError(errorTodo, err.Error())
		return attr
	}

	attr.fileSize = uint32(fileInfo.Size())
	attr.fileType = osFileFound
	if fileInfo.IsDir() {
		attr.fileType = osDirectoryFound
	}

	/*
		Search metadata file "{filename}.inf" looking like:
		$.BasObj     003000 003100 005000 00 CRC32=614721E1
	*/
	attr.hasMetadata = false
	data, err := os.ReadFile(filename + ".inf")
	if errors.Is(err, os.ErrNotExist) {
		return attr
	}
	parts := strings.Fields(string(data))
	if len(parts) < 3 {
		env.log(fmt.Sprintf("Invalid format for metadata file %s.inf, missing fields", filename))
		return attr
	}

	i, err := strconv.ParseInt(parts[1], 16, 32)
	if err != nil {
		env.log(fmt.Sprintf("Invalid format for metadata file %s.inf, bad load address", filename))
		return attr
	}
	attr.loadAddress = uint32(i)

	i, err = strconv.ParseInt(parts[2], 16, 32)
	if err != nil {
		env.log(fmt.Sprintf("Invalid format for metadata file %s.inf, bad exec address", filename))
		return attr
	}
	attr.executionAddress = uint32(i)

	attr.hasMetadata = true
	return attr
}
