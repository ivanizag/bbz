package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
)

/*
	Start commands. The order is relevant to find the proper shortcuts.
	Example: *L. is *LOAD and not *LINE
	See:
		https://github.com/raybellis/mos120/blob/2e2ff80708e79553643e4b77c947b0652117731b/mos120.s#L6839
		BBC Microcomputer Advanced User Guide, chapter 2.
*/
var cliCommands = []string{
	"CAT",
	"FX",
	"BASIC",
	"BYE",
	"CODE",
	"DIR",
	"DELETE",
	"DRIVE",
	"EXEC",
	"EX",
	"HELP",
	"HOST", // Added for bbz
	"INFO",
	"KEY",
	"LOAD",
	"LINE",
	"MOTOR",
	"OPT",
	"QUIT", // Added for bbz
	"RUN",
	"ROM",
	"ROMS",
	"SAVE",
	"SPOOL",
	"TAPE",
	"TV",
	"TYPE",
}

func execOSCLI(env *environment) {
	a, x, y, p := env.cpu.GetAXYP()

	/*
		This routine passes a line of text to the command line
		interpreter which decodes and executes any command
		recognised.
		X and Y should point to a line of text terminated by a
		carriage return character (ASCII &0D/13)
	*/
	xy := uint16(x) + uint16(y)<<8
	lineNotTerminated := env.mem.peekString(xy, 0x0d)
	line := lineNotTerminated + "\r"
	pos := 0
	pos = parseSkipSpaces(line, pos)
	if line[pos] == '*' { // Skip '*' if present
		pos++
	}
	pos = parseSkipSpaces(line, pos)
	if line[pos] == '|' || line[pos] == '\r' { // Ignore "*|" or standalone "*"
		env.log(fmt.Sprintf("OSCLI('%s', CMD=empty)", lineNotTerminated))
		return
	}
	command := ""
	if line[pos] == '/' { // Send "*/[...]" to filling system
		command = "RUN"
		pos++
	} else {
		// Extract command
		for ; !strings.ContainsAny(string(line[pos]), " .0123456789\r\""); pos++ {
			command += string(line[pos])
		}
		command = strings.ToUpper(command) // Commands are case insensitive
		if line[pos] == '.' {
			// Expand . shortcut
			for _, candidate := range cliCommands {
				if strings.HasPrefix(candidate, command) {
					command = candidate // Full command found
					break
				}
			}
			pos++
		}
	}
	pos = parseSkipSpaces(line, pos)
	unhandled := false
	valid := false

	env.log(fmt.Sprintf("OSCLI('%s', CMD='%s', ARGS='%s')", lineNotTerminated, command, lineNotTerminated[pos:]))

	switch command {

	case "FX":
		var argA uint8
		pos, argA, valid = parseByte(line, pos)
		if !valid {
			env.raiseError(254, "Bad Command")
			break
		}
		execOSCLIfx(env, uint8(argA), line, pos)

	case "BASIC":
		// Runs the first language ROM with no service entry
		unhandled = true
		for slot := 0xf; slot >= 0; slot-- {
			romType := env.mem.data[mosRomTypeTable+uint16(slot)]
			if romType&0b1100_0000 == 0b0100_0000 { // bit 7 clear, bit 6 set
				env.initLanguage(uint8(slot))
				unhandled = false
				break
			}
		}

	case "CAT":
		pathName := ""
		_, pathName, valid = parseFilename(line, pos)
		if !valid {
			env.raiseError(253, "Bad String")
			break
		}
		if pathName == "" || pathName == "$" {
			pathName = "."
		}

		entries, err := os.ReadDir(pathName)
		if err != nil {
			env.raiseError(errorTodo, err.Error())
			break
		}

		for _, entry := range entries {
			if strings.HasPrefix(entry.Name(), ".") ||
				strings.HasSuffix(entry.Name(), metadataExtension) {
				// Ignore the file
			} else if entry.Type().IsRegular() {
				env.con.writef("%v\n", entry.Name())
			} // else ignore the file
		}

	case "EX":
		pathName := ""
		_, pathName, valid = parseFilename(line, pos)
		if !valid {
			env.raiseError(253, "Bad String")
			break
		}
		if pathName == "" || pathName == "$" {
			pathName = "."
		}

		entries, err := os.ReadDir(pathName)
		if err != nil {
			env.raiseError(errorTodo, err.Error())
			break
		}

		for _, entry := range entries {
			if entry.Type().IsRegular() {
				fullName := path.Join(pathName, entry.Name())
				attr := getFileAttributes(env, fullName)
				env.con.writef("%-22v %06x %06x %06x \n",
					entry.Name(),
					attr.loadAddress&0xff_ffff,
					attr.executionAddress&0xff_ffff,
					attr.fileSize&0xff_ffff)
			}
		}

	case "CODE":
		execOSCLIfx(env, 0x88, line, pos)

	case "DELETE":
		// *DELETE filename
		filename := ""
		_, filename, valid = parseFilename(line, pos)
		if !valid {
			env.raiseError(253, "Bad String")
			break
		}

		if filename != "" {
			deleteFile(env, filename)
		}

	case "DIR":
		path := ""
		_, path, valid = parseFilename(line, pos)
		if !valid {
			env.raiseError(253, "Bad String")
			break
		}

		if path == "" || path == "$" {
			break
			/*
				var err error
				path, err = os.UserHomeDir()
				if err != nil {
					env.raiseError(errorTodo, err.Error())
					break
				}

			*/
		}

		err := os.Chdir(path)
		if err != nil {
			env.raiseError(206, "Bad directory")
			break
		}

	case "DRIVE":
		var drive uint8
		_, drive, valid = parseByte(line, pos)
		if !valid || drive >= 4 {
			env.raiseError(205, "Bad drive")
		}
		// Do nothing. We could use subdirs for drive 1, 2 and 3

	//case "EXEC":

	case "HELP":
		env.con.write("\nBBZ 0.0\n")

		// Send to the other ROMS if available.
		env.mem.pokeWord(zpStr, xy)
		env.cpu.SetAXYP(a, serviceHELP, uint8(pos), p)
		env.cpu.SetPC(procOSBYTE_143)

	case "HOST":
		args := line[pos:]
		args = strings.TrimSuffix(args, "\r")
		if len(args) == 0 {
			env.raiseError(errorTodo, "Command missing for *HOST")
			break
		}
		params := strings.Split(args, " ")
		cmd := exec.Command(params[0], params[1:]...)
		stdout, err := cmd.Output()
		if err != nil {
			env.raiseError(errorTodo, err.Error())
		}
		env.con.write(string(stdout))
		env.con.write("\n")

	case "INFO":
		filename := ""
		_, filename, valid = parseFilename(line, pos)
		if !valid {
			env.raiseError(253, "Bad String")
			break
		}

		attr := getFileAttributes(env, filename)
		if attr.hasMetadata {
			env.con.writef("%s\t %06X %06X %06X\n", filename,
				attr.loadAddress, attr.executionAddress, attr.fileSize)
		} else {
			env.con.writef("%s\t ?????? ?????? %06X\n", filename, attr.fileSize)
		}

	// case "KEY":
	case "LOAD":
		// *LOAD <filename> [<address>]
		filename := ""
		pos, filename, valid = parseFilename(line, pos)
		if !valid {
			env.raiseError(253, "Bad String")
			break
		}

		loadAddress := addressNull
		if line[pos] != '\r' {
			_, loadAddress, valid = parseDWord(line, pos)
			if !valid {
				env.raiseError(252, "Bad address")
				break
			}
		}
		loadFile(env, filename, loadAddress)

	// case "LINE":
	case "MOTOR":
		execOSCLIfx(env, 0x89, line, pos)
	case "OPT":
		execOSCLIfx(env, 0x8b, line, pos)
	case "BYE":
		fallthrough
	case "QUIT":
		env.stop = true
	case "RUN":
		// *RUN <filename>
		filename := ""
		_, filename, valid = parseFilename(line, pos)
		if !valid {
			env.raiseError(253, "Bad String")
			break
		}

		attr := loadFile(env, filename, addressNull)
		if attr.fileType == osFileFound {
			if attr.hasMetadata {
				env.cpu.SetPC(uint16(attr.executionAddress))
			} else {
				env.raiseError(errorTodo, "Missing metadata file")
			}
		}

	case "ROM":
		execOSCLIfx(env, 0x8d, line, pos)

	case "ROMS":
		currentRom := env.mem.Peek(sheilaRomLatch)
		for i := 0xf; i >= 0; i-- {
			if env.mem.writeProtectRom[i] {
				env.mem.Poke(sheilaRomLatch, uint8(i))
				name := env.mem.peekString(romTitleString, 0)
				if name == "" {
					env.con.writef("ROM %X ?\n", i)
				} else {
					version := env.mem.Peek(romVersion)
					romType := env.mem.Peek(romTypeByte)
					attributes := "("
					if romType&0x80 != 0 {
						attributes += "S"
					}
					if romType&0x40 != 0 {
						attributes += "L"
					}
					attributes += ")"

					env.con.writef("ROM %X %s %02v %s\n", i, name, version, attributes)
				}
			} else {
				env.con.writef("RAM %X 16K\n", i)
			}
		}
		env.mem.Poke(sheilaRomLatch, currentRom)

	case "SAVE":
		// *SAVE <filename> <start addr> <end addr or length> [<exec addr>] [<reload addr>]
		// *SAVE A 1a34+2a
		filename := ""
		pos, filename, valid = parseFilename(line, pos)
		if !valid {
			env.raiseError(253, "Bad String")
			break
		}

		var startAddress uint32
		pos, startAddress, valid = parseDWord(line, pos)
		if !valid {
			env.raiseError(252, "Bad address")
			break
		}

		isSize := false
		if line[pos] == '+' {
			isSize = true
			pos++
		}
		var endAddress uint32
		pos, endAddress, valid = parseDWord(line, pos)
		if !valid {
			env.raiseError(252, "Bad address")
			break
		}
		if isSize {
			endAddress = startAddress + endAddress
		}

		executionAddress := startAddress
		if line[pos] != '\r' {
			pos, executionAddress, valid = parseDWord(line, pos)
			if !valid {
				env.raiseError(252, "Bad address")
				break
			}
		}

		loadAddress := startAddress
		if line[pos] != '\r' {
			pos, loadAddress, valid = parseDWord(line, pos)
			if !valid {
				env.raiseError(252, "Bad address")
				break
			}
		}

		if line[pos] != '\r' {
			env.raiseError(254, "Bad command")
			break
		}

		saveFile(env, filename, startAddress, endAddress, executionAddress, loadAddress, false)

	case "SPOOL":
		// *SPOOL filename
		// *SPOOL
		spoolFile := env.mem.Peek(mosSpoolFileHandle)
		if spoolFile != 0 {
			env.closeFile(spoolFile)
			env.mem.Poke(mosSpoolFileHandle, 0)
		}

		filename := ""
		_, filename, valid = parseFilename(line, pos)
		if !valid {
			env.raiseError(253, "Bad String")
			break
		}
		if filename != "" {
			// Activate spool
			spoolFile := env.openFile(filename, 0x80 /*open for output*/)
			env.mem.Poke(mosSpoolFileHandle, spoolFile)
		}

	case "TAPE":
		execOSCLIfx(env, 0x8c, line, pos)
	case "TV":
		execOSCLIfx(env, 0x90, line, pos)

	case "TYPE":
		// *TYPE filename
		filename := ""
		_, filename, valid = parseFilename(line, pos)
		if !valid {
			env.raiseError(253, "Bad String")
			break
		}

		if filename != "" {
			data, err := os.ReadFile(filename)
			if err != nil {
				env.raiseError(errorTodo, err.Error())
				break
			}
			for _, ch := range data {
				env.vdu.write(ch)
			}
		}

	default:
		unhandled = true
	}

	if unhandled {
		// Send to the other ROMS if available.
		env.mem.pokeWord(zpStr, xy)
		env.cpu.SetAXYP(serviceOSCLI, x, 1, p)
		env.cpu.SetPC(procServiceRoms)
		// procServiceRoms issues a 254-Bad command if the command is not handled by any ROM
	}
}

func execOSCLIfx(env *environment, argA uint8, line string, pos int) {
	argX := uint8(0)
	argY := uint8(0)
	var valid bool

	if line[pos] != '\r' {
		if line[pos] == ',' {
			pos++ // Skip ','
		}
		pos = parseSkipSpaces(line, pos)
		pos, argX, valid = parseByte(line, pos)
		if !valid {
			env.raiseError(254, "Bad Command")
			return
		}
	}

	if line[pos] != '\r' {
		if line[pos] == ',' {
			pos++ // Skip ','
		}
		pos = parseSkipSpaces(line, pos)
		_, argY, valid = parseByte(line, pos)
		if !valid {
			env.raiseError(254, "Bad Command")
			return
		}
	}

	// Send to OSBYTE
	_, _, _, p := env.cpu.GetAXYP()
	env.cpu.SetAXYP(uint8(argA), uint8(argX), uint8(argY), p)
	execOSBYTE(env)
}

func parseFilename(line string, pos int) (int, string, bool) {
	terminator := uint8(' ')
	if line[pos] == '"' {
		terminator = '"'
		pos++
	}
	cursor := pos

	for line[cursor] != terminator && line[cursor] != '\r' {
		cursor++
	}

	filename := line[pos:cursor]
	if terminator != ' ' {
		if line[cursor] != terminator {
			return cursor, "", false
		}
		cursor++
	}
	cursor = parseSkipSpaces(line, cursor)

	return cursor, filename, true
}

func parseSkipSpaces(line string, pos int) int {
	for line[pos] == ' ' { // Remove initial spaces
		pos++
	}
	return pos
}

func parseByte(line string, pos int) (int, uint8, bool) {
	cursor := pos
	for line[cursor] >= '0' && line[cursor] <= '9' {
		cursor++
	}
	if cursor == pos {
		return pos, 0, false
	}
	value, err := strconv.Atoi(line[pos:cursor])
	if err != nil || value > 255 {
		return cursor, 0, false
	}

	cursor = parseSkipSpaces(line, cursor)
	return cursor, uint8(value), true
}

func parseDWord(line string, pos int) (int, uint32, bool) {
	cursor := pos
	for (line[cursor] >= '0' && line[cursor] <= '9') ||
		(line[cursor] >= 'a' && line[cursor] <= 'f') ||
		(line[cursor] >= 'A' && line[cursor] <= 'F') {
		cursor++
	}
	if cursor == pos {
		return pos, 0, false
	}
	value, err := strconv.ParseUint(line[pos:cursor], 16, 32)
	if err != nil {
		return cursor, 0, false
	}

	cursor = parseSkipSpaces(line, cursor)
	return cursor, uint32(value), true
}
