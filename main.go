package main

import (
	"flag"
	"fmt"
)

func main() {
	fmt.Printf("bbz - Acorn MOS for 6502 adaptation layer, https://github.com/ivanizag/bbz\n")
	fmt.Printf("(tip: uppercase is usually needed)\n\n")

	traceCPU := flag.Bool(
		"c",
		false,
		"dump to the console the CPU execution operations")
	traceMOS := flag.Bool(
		"m",
		false,
		"dump to the console the MOS calls excluding console I/O calls")
	traceMOSFull := flag.Bool(
		"M",
		false,
		"dump to the console the MOS calls including console I/O calls")
	traceMemory := flag.Bool(
		"s",
		false,
		"dump to the console the accesses to Fred, Jim or Sheila")
	panicOnErr := flag.Bool(
		"p",
		false,
		"panic on not implemented MOS calls")

	flag.Parse()

	romFile := flag.Arg(0)
	if romFile == "" {
		romFile = "BASIC.ROM"
	}

	RunMOSEnvironment(romFile, "firmware",
		*traceCPU,
		(*traceMOS) || (*traceMOSFull),
		*traceMOSFull,
		*traceMemory,
		*panicOnErr)
}
