package main

import (
	"fmt"
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
	line := env.getStringFromMem(xy, 0x0d)
	fields := strings.Fields(line)
	command := fields[0]
	params := fields[1:]

	if strings.HasPrefix(command, "*FX") {
		// *FX123 should be treated as *FX 123
		fxNumber := command[3:]
		if len(fxNumber) != 0 {
			command = "*FX"
			params = append([]string{fxNumber}, params...)
		}
	}

	msg := "\nBad command\n"
	switch command {
	case "*HELP":
		msg = fmt.Sprintf("\nbbz - Acorn MOS for 6502 adaptation layer, https://github.com/ivanizag/bbz\n")
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
		msg = ""
	}

	fmt.Print(msg)
	env.log(fmt.Sprintf("OSCLI('%s %s)", command, strings.Join(params, " ")))
}
