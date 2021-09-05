package main

import (
	"fmt"
	"os"
	"os/exec"
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
	"CODE",
	"DIR",
	"EXEC",
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
	lineNotTerminated := env.mem.getString(xy, 0x0d)
	line := lineNotTerminated + "\r"
	pos := 0

	for line[pos] == ' ' { // Remove initial spaces
		pos++
	}
	if line[pos] == '*' { // Skip '*' if present
		pos++
	}
	for line[pos] == ' ' { // Remove spaces after '*'
		pos++
	}
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
	for line[pos] == ' ' { // Remove spaces after command
		pos++
	}
	args := line[pos:]
	args = strings.TrimRight(args, " \r")
	unhandled := false

	env.log(fmt.Sprintf("OSCLI('%s', CMD='%s')", lineNotTerminated, command))

	switch command {
	case "CAT":
		// TODO
		fmt.Println("\n<<directory placeholder>>")

	case "FX":
		params := strings.Split(args, ",")
		if len(params) == 0 || len(params) > 3 {
			env.raiseError(254, "Bad Command")
			break
		}
		argA, err := strconv.Atoi(strings.TrimSpace(params[0]))
		if err != nil || (argA&0xff) != argA {
			env.raiseError(254, "Bad Command")
			break
		}
		execOSCLIfx(env, uint8(argA), params[1:])

	case "BASIC":
		// Runs the first language ROM with no service entry
		unhandled = true
		for slot := 0xf; slot >= 0; slot-- {
			romType := env.mem.data[romTypeTable+uint16(slot)]
			if romType&0b1100_0000 == 0b0100_0000 { // bit 7 clear, bit 6 set
				env.initLanguage(uint8(slot))
				unhandled = false
				break
			}
		}

	case "CODE":
		execOSCLIfx(env, 0x88, strings.Split(args, ","))

	case "DIR":
		dest := args
		if len(dest) == 0 {
			var err error
			dest, err = os.UserHomeDir()
			if err != nil {
				env.raiseError(errorTodo, err.Error())
				break
			}
		}

		err := os.Chdir(dest)
		if err != nil {
			env.raiseError(206, "Bad directory")
			break
		}

	//case "EXEC":

	case "HELP":
		fmt.Println("\nBBZ 0.0")

		// Send to the other ROMS if available.
		env.mem.pokeWord(zpStr, xy)
		env.cpu.SetAXYP(a, serviceHELP, uint8(pos), p)
		env.cpu.SetPC(procOSBYTE_143)

	case "HOST":
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
		fmt.Println(string(stdout))

	case "INFO":
		attr := getFileAttributes(env, args)
		if attr.hasMetadata {
			fmt.Printf("%s\t %06X %06X %06X\n", args, attr.loadAddress, attr.executionAddress, attr.fileSize)
		} else {
			fmt.Printf("%s\t ?????? ?????? %06X\n", args, attr.fileSize)
		}

	// case "KEY":
	case "LOAD":
		// *LOAD <filename> [<address>]
		params := strings.Split(args, " ")
		if len(params) > 2 {
			env.raiseError(254, "Bad command")
			break
		}
		if len(params) == 0 {
			env.raiseError(214, "File not found")
			break
		}
		filename := cleanFilename(params[0])
		loadAddress := loadAddressNull
		if len(params) >= 2 {
			i, err := strconv.ParseInt(params[1], 16, 32)
			if err != nil {
				env.raiseError(252, "Bad address")
				break
			}
			loadAddress = uint32(i)
		}
		loadFile(env, filename, loadAddress)

	// case "LINE":
	case "MOTOR":
		execOSCLIfx(env, 0x89, strings.Split(args, ","))
	case "OPT":
		execOSCLIfx(env, 0x8b, strings.Split(args, ","))
	case "QUIT":
		env.stop = true
	case "RUN":
		// *RUN <filename>
		params := strings.Split(args, " ")
		if len(params) == 0 {
			env.raiseError(214, "File not found")
			break
		}
		filename := cleanFilename(params[0])
		attr := loadFile(env, filename, loadAddressNull)
		if attr.fileType == osFileFound {
			if attr.hasMetadata {
				env.cpu.SetPC(uint16(attr.executionAddress))
			} else {
				env.raiseError(errorTodo, "Missing metadata file")
			}
		}

	case "ROM":
		execOSCLIfx(env, 0x8d, strings.Split(args, ","))

	case "ROMS":
		selectedROM := env.mem.Peek(sheilaRomLatch)
		for i := 0xf; i >= 0; i-- {
			if env.mem.writeProtectRom[i] {
				env.mem.Poke(sheilaRomLatch, uint8(i))
				name := env.mem.getString(romTitleString, 0)
				if name == "" {
					fmt.Printf("ROM %X ?\n", i)
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

					fmt.Printf("ROM %X %s %02v %s\n", i, name, version, attributes)
				}
			} else {
				fmt.Printf("RAM %X 16K\n", i)
			}
		}
		env.mem.Poke(sheilaRomLatch, selectedROM)

	case "SAVE":
		// *SAVE <filename> <start addr> <end addr or length> <exec addr> <reload addr>
		params := strings.Split(args, " ")
		if len(params) < 3 || len(params) > 5 {
			env.raiseError(254, "Bad command")
			break
		}

		filename := cleanFilename(params[0])

		i, err := strconv.ParseInt(params[1], 16, 32)
		if err != nil {
			env.raiseError(252, "Bad address")
			break
		}
		startAddress := uint32(i)

		isSize := false
		if params[2][0] == '+' {
			isSize = true
			params[2] = params[2][1:]
		}
		i, err = strconv.ParseInt(params[2], 16, 32)
		if err != nil {
			env.raiseError(252, "Bad address")
			break
		}
		endAddress := uint32(i)
		if isSize {
			endAddress = startAddress - endAddress
		}

		if len(params) > 3 {
			env.notImplemented("*SAVE with execution o reload address")
		}

		saveFile(env, filename, startAddress, endAddress)

	// case "SPOOL":
	case "TAPE":
		execOSCLIfx(env, 0x8c, strings.Split(args, ","))
	case "TV":
		execOSCLIfx(env, 0x90, strings.Split(args, ","))

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

func execOSCLIfx(env *environment, argA uint8, params []string) {
	argX := 0
	if len(params) >= 1 {
		var err error
		argX, err = strconv.Atoi(strings.TrimSpace(params[0]))
		if err != nil || (argX&0xff) != argX {
			env.raiseError(254, "Bad Command")
			return
		}
	}
	argY := 0
	if len(params) >= 2 {
		var err error
		argY, err = strconv.Atoi(strings.TrimSpace(params[1]))
		if err != nil || (argY&0xff) != argY {
			env.raiseError(254, "Bad Command")
			return
		}
	}

	// Send to OSBYTE
	_, _, _, p := env.cpu.GetAXYP()
	env.cpu.SetAXYP(uint8(argA), uint8(argX), uint8(argY), p)
	execOSBYTE(env)
}

func cleanFilename(filename string) string {
	if filename[0] == '"' && len(filename) >= 2 {
		return filename[1:(len(filename) - 1)]
	}
	return filename
}
