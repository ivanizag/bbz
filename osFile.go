package main

import (
	"errors"
	"fmt"
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

	addressNull uint32 = ^uint32(0)
)

type fileAttributes struct {
	fileType         uint8
	fileSize         uint32
	hasMetadata      bool
	loadAddress      uint32
	executionAddress uint32
	attributes       uint32
}

func execOSFILE(env *environment) {
	a, x, y, p := env.cpu.GetAXYP()

	// See: http://beebwiki.mdfs.net/OSFILE
	controlBlock := uint16(x) + uint16(y)<<8
	filenameAddress := env.mem.peekWord(controlBlock)
	loadAddress := env.mem.peekDoubleWord(controlBlock + cbLoadAddress)
	executionAddress := env.mem.peekDoubleWord(controlBlock + cbExecutionAddress)
	startAddress := env.mem.peekDoubleWord(controlBlock + cbStartAddressOrSize)
	endAddress := env.mem.peekDoubleWord(controlBlock + cbEndAddressOrAttributes)

	filename := env.mem.peekString(filenameAddress, 0x0d)

	newA := uint8(0) // Nothing found
	option := ""
	switch a {
	case 0:
		option = "Save file"
		/*
			Save a block of memory as a file using the
			information provided in the parameter block.
		*/
		attr := saveFile(env, filename, startAddress, endAddress, executionAddress, loadAddress, false)
		if attr.fileType != osNotFound {
			updateControlBlock(env, controlBlock, attr)
		}
		newA = attr.fileType

	case 1:
		option = "Write file metadata"
		attr := getFileAttributes(env, filename)
		if attr.fileType != osNotFound {
			attr.loadAddress = loadAddress
			attr.executionAddress = executionAddress
			attr.attributes = endAddress
			writeMetada(env, filename, attr)
		}
		newA = attr.fileType

	case 2:
		option = "Write file reload address"
		attr := getFileAttributes(env, filename)
		if attr.fileType != osNotFound {
			attr.loadAddress = loadAddress
			writeMetada(env, filename, attr)
		}
		newA = attr.fileType

	case 3:
		option = "Write file execution address"
		attr := getFileAttributes(env, filename)
		if attr.fileType != osNotFound {
			attr.executionAddress = executionAddress
			writeMetada(env, filename, attr)
		}
		newA = attr.fileType

	case 4:
		option = "Write file attributes"
		attr := getFileAttributes(env, filename)
		if attr.fileType != osNotFound {
			attr.attributes = endAddress
			writeMetada(env, filename, attr)
		}
		newA = attr.fileType

	case 5:
		option = "File info"
		/*
			Read a file’s catalogue information, with the file
			type returned in the accumulator. The information
			is written to the parameter block.
		*/
		attr := getFileAttributes(env, filename)
		if attr.fileType != osNotFound {
			updateControlBlock(env, controlBlock, attr)
		}
		newA = attr.fileType

	case 6:
		option = "Delete file"
		/*
			Delete object. If the object does not exist, A returned
			as &00. If the object is locked, or is not owned, or is
			a directory that is not empty, or is open, then an error
			is generated.
		*/
		newA = deleteFile(env, filename)

	case 7:
		option = "Create an empty file of defined length"
		/*
			Create an empty file of defined length. Block as for
			SAVE. The supplied start address is usually passed as &0
			and the end address as the required length. No data is
			transfered, and the file does not necessarily contain
			zeros. Some file systems may deliberately overwrite any
			existing data in the file. If a file already exists with
			the same name, it is overwritten, with the file access
			and the case of the name staying the same. If the file
			is locked, or a directory exists with the same name, or
			the file is open, then an error is generated.
		*/
		attr := saveFile(env, filename, startAddress, endAddress, executionAddress, loadAddress, true)
		if attr.fileType != osNotFound {
			updateControlBlock(env, controlBlock, attr)
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
			loadAddress = addressNull
		}

		attr := loadFile(env, filename, loadAddress)
		//env.mem.pokeDoubleWord(controlBlock+cbStartAddressOrSize, attr.fileSize)
		if attr.fileType == osNotFound {
			env.raiseError(214, "File not found")
		}
		updateControlBlock(env, controlBlock, attr)
		newA = attr.fileType

	default:
		env.notImplemented(fmt.Sprintf("OSFILE(A=%02x)", a))
	}

	if a != 1 && a != 5 {
		env.mem.pokeWord(controlBlock+cbEndAddressOrAttributes, 0x00 /*attributes*/)
	}
	env.cpu.SetAXYP(newA, x, y, p)
	env.log(fmt.Sprintf("OSFILE('%s',A=%02x,FCB=%04x,FILE=%s) => %v",
		option, a, controlBlock, filename, newA))

}

func updateControlBlock(env *environment, controlBlock uint16, attr *fileAttributes) {
	if attr.fileType != osNotFound {
		env.mem.pokeDoubleWord(controlBlock+cbStartAddressOrSize, attr.fileSize)
		if attr.hasMetadata {
			env.mem.pokeDoubleWord(controlBlock+cbLoadAddress, attr.loadAddress)
			env.mem.pokeDoubleWord(controlBlock+cbExecutionAddress, attr.executionAddress)
			env.mem.pokeDoubleWord(controlBlock+cbEndAddressOrAttributes, attr.attributes)
		}
	}
}

func loadFile(env *environment, filename string, loadAddress uint32) *fileAttributes {
	attr := getFileAttributes(env, filename)
	if attr.fileType == osNotFound {
		return attr
	}
	if attr.fileType == osDirectoryFound {
		env.raiseError(errorTodo, "Directory found")
		return attr
	}

	if loadAddress == addressNull {
		if !attr.hasMetadata {
			env.raiseError(errorTodo, "Missing metadata file")
			attr.fileType = osNotFound
			return attr
		}
		loadAddress = attr.loadAddress
	} else {
		attr.loadAddress = loadAddress
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		env.raiseError(errorTodo, err.Error())
		attr.fileType = osNotFound
		return attr
	}

	env.mem.pokeSlice(uint16(loadAddress), uint16(len(data)), data)
	return attr
}

func saveFile(env *environment, filename string,
	startAddress uint32, endAddress uint32, executionAddress uint32, loadAddress uint32, blank bool) *fileAttributes {

	var attr fileAttributes
	attr.loadAddress = loadAddress
	attr.executionAddress = executionAddress
	attr.fileSize = endAddress - startAddress
	// attr.attributes = ?

	var data []uint8
	if blank {
		data = make([]uint8, attr.fileSize)
	} else {
		data = env.mem.peekSlice(uint16(startAddress), uint16(attr.fileSize))
	}

	err := os.WriteFile(filename, data, 0644)
	if err != nil {
		env.raiseError(errorTodo, err.Error())
		attr.fileType = osNotFound
	} else {
		attr.fileType = osFileFound
	}

	writeMetada(env, filename, &attr)
	return &attr
}

func deleteFile(env *environment, filename string) uint8 {
	err := os.Remove(filename)
	var pathError *os.PathError
	if errors.As(err, &pathError) {
		return osNotFound
	}
	if err != nil {
		env.raiseError(errorTodo, err.Error())
		return osNotFound
	}

	os.Remove(filename + ".inf")
	return osFileFound
}

func getFileAttributes(env *environment, filename string) *fileAttributes {
	var attr fileAttributes

	fileInfo, err := os.Stat(filename)
	if errors.Is(err, os.ErrNotExist) {
		attr.fileType = osNotFound
		return &attr
	}
	if err != nil {
		attr.fileType = osNotFound
		return &attr
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
		return &attr
	}
	parts := strings.Fields(string(data))
	if len(parts) < 5 {
		env.log(fmt.Sprintf("Invalid format for metadata file %s.inf, missing fields", filename))
		return &attr
	}

	i, err := strconv.ParseUint(parts[1], 16, 64)
	if err != nil {
		env.log(fmt.Sprintf("Invalid format for metadata file %s.inf, bad load address '%s'", filename, err.Error()))
		return &attr
	}
	attr.loadAddress = uint32(i)

	i, err = strconv.ParseUint(parts[2], 16, 64)
	if err != nil {
		env.log(fmt.Sprintf("Invalid format for metadata file %s.inf, bad exec address '%s'", filename, err.Error()))
		return &attr
	}
	attr.executionAddress = uint32(i)

	i, err = strconv.ParseUint(parts[4], 16, 64)
	if err != nil {
		env.log(fmt.Sprintf("Invalid format for metadata file %s.inf, bad sttributes '%s'", filename, err.Error()))
		return &attr
	}
	attr.attributes = uint32(i)

	attr.hasMetadata = true
	return &attr
}

func writeMetada(env *environment, filename string, attr *fileAttributes) {
	// $.BasObj     003000 003100 005000 00 CRC32=614721E1
	metadata := fmt.Sprintf("$.FILE    %08X %08X %08X %02X",
		attr.loadAddress, attr.executionAddress, attr.fileSize, attr.attributes)
	os.WriteFile(filename+".inf", []byte(metadata), 0644)
}
