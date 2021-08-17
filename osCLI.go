package main

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

func execOSCLI(env *environment) {
	_, x, y, p := env.cpu.GetAXYP()

	/*
		This routine passes a line of text to the command line
		interpreter which decodes and executes any command
		recognised.
		X and Y should point to a line of text terminated by a
		carriage return character (ASCII &0D/13)
	*/
	xy := uint16(x) + uint16(y)<<8
	line := env.mem.getString(xy, 0x0d)
	fields := strings.Fields(line)
	command := strings.ToUpper(fields[0])
	// The command-line interpreter does not distinguish between upper and lower case characters in the command name
	command = strings.ToUpper(command)
	params := fields[1:]
	handled := false

	if command == "*" && len(params) > 0 {
		// There are spaces between the * and the command
		command = "*" + strings.ToUpper(params[0])
		params = params[1:]
	}

	if strings.HasPrefix(command, "*FX") {
		// *FX123 should be treated as *FX 123
		fxNumber := command[3:]
		if len(fxNumber) != 0 {
			command = "*FX"
			params = append([]string{fxNumber}, params...)
		}
	}

	if strings.HasPrefix(command, "*|") {
		// *|123 should be treated as *| 123
		fxNumber := command[3:]
		if len(fxNumber) != 0 {
			command = "*|"
			params = append([]string{fxNumber}, params...)
		}
	}

	msg := ""
	switch command {

	case "*|":
		/*
			An operating system command-line with a ‘|’, string escape
			character, as its first non-blank character will be ignored by the
			operating system. This could be used to put comment lines into
			a series of operating system commands placed in an EXEC file
			for example.
		*/
		// do nothing
		handled = true

	case "*H.":
		fallthrough
	case "*HELP":
		msg = "bbz - Acorn MOS for 6502 adaptation layer, https://github.com/ivanizag/bbz"
		handled = true

		// TODO: multiple ROMS: service call 9 after the MOS message

	case "*.":
		fallthrough
	case "*CAT":
		// TODO
		msg = "<<directory placeholder>>"
		handled = true

	case "*QUIT":
		env.stop = true
		handled = true

	case "*HOST":
		if len(params) == 0 {
			env.raiseError(errorTodo, "Command missing for *HOST")
		} else {
			cmd := exec.Command(params[0], params[1:]...)
			stdout, err := cmd.Output()
			if err != nil {
				env.raiseError(errorTodo, err.Error())
			}
			fmt.Println(string(stdout))
		}
		handled = true

	case "*FX":
		// Parse  *FX args
		if len(params) == 0 || len(params) > 3 {
			break
		}
		argA, err := strconv.Atoi(params[0])
		if err != nil || (argA&0xff) != argA {
			break
		}
		argX := 0
		if len(params) >= 2 {
			argX, err = strconv.Atoi(params[1])
			if err != nil || (argX&0xff) != argX {
				break
			}
		}
		argY := 0
		if len(params) >= 3 {
			argY, err = strconv.Atoi(params[2])
			if err != nil || (argY&0xff) != argY {
				break
			}
		}

		// Send to OSBYTE
		env.cpu.SetAXYP(uint8(argA), uint8(argX), uint8(argY), p)
		execOSBYTE(env)
		handled = true

	case "*BASIC":
		// Runs the first language ROM with no service entry
		for slot := 0xf; slot >= 0; slot-- {
			romType := env.mem.data[romTypeTable+uint16(slot)]
			if romType&0b1100_0000 == 0b0100_0000 { // bit 7 clear, bit 6 set
				env.initLanguage(uint8(slot))
				handled = true
				break
			}
		}
	}

	if !handled {
		// Send to the other ROMS if available.
		env.mem.pokeWord(zpStr, xy)
		cmd := uint8(4) // Unrecognized command
		env.cpu.SetAXYP(cmd, x, 1, p)
		env.cpu.SetPC(procCLITOROMS)
		// procCLITOROMS issues a 254-Bad command if the command is not handled by any ROM
	}

	if msg != "" {
		fmt.Printf("\n%s\n", msg)
	}
	env.log(fmt.Sprintf("OSCLI('%s %s')", command, strings.Join(params, " ")))
}
