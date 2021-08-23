package main

import (
	"fmt"
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
	"EXEC",
	"HELP",
	"HOST", // Added for bbz
	"KEY",
	"LOAD",
	"LINE",
	"MOTOR",
	"OPT",
	"QUIT", // Added for bbz
	"RUN",
	"ROM",
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
		return
	}
	if line[pos] == '/' { // Send "*/[...]" to filling system
		env.log("*/[...] not implemented")
		return
	}

	// Extract command
	command := ""
	for ; !strings.ContainsAny(string(line[pos]), " .0123456789\r"); pos++ {
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
	//case "EXEC":
	case "HELP":
		fmt.Println("\nbbz - Acorn MOS for 6502 adaptation layer, https://github.com/ivanizag/bbz")

		// Send to the other ROMS if available.
		env.mem.pokeWord(zpStr, xy)
		env.cpu.SetAXYP(a, serviceHELP, 1, p)
		env.cpu.SetPC(procOSBYTE_143)

	case "HOST":
		if len(args) == 0 {
			env.raiseError(errorTodo, "Command missing for *HOST")
		} else {
			params := strings.Split(args, " ")
			cmd := exec.Command(params[0], params[1:]...)
			stdout, err := cmd.Output()
			if err != nil {
				env.raiseError(errorTodo, err.Error())
			}
			fmt.Println(string(stdout))
		}

	// case "KEY":
	// case "LOAD":
	// case "LINE":
	case "MOTOR":
		execOSCLIfx(env, 0x89, strings.Split(args, ","))
	case "OPT":
		execOSCLIfx(env, 0x8b, strings.Split(args, ","))
	case "QUIT":
		env.stop = true
	// case "RUN":
	case "ROM":
		execOSCLIfx(env, 0x8d, strings.Split(args, ","))
	// case "SAVE":
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
		cmd := uint8(serviceOSCLI)
		env.cpu.SetAXYP(cmd, x, 1, p)
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
