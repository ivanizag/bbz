package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
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
	rawline := flag.Bool(
		"r",
		false,
		"disable readline like input with history")
	roms := make([]*string, 16)
	for i := 0; i < 16; i++ {
		roms[i] = flag.String(
			fmt.Sprintf("rom%v", i),
			"",
			fmt.Sprintf("filename for rom %v (slot 0x%x)", i, 15-i))
	}

	flag.Parse()

	if *roms[0] == "" {
		romFile := flag.Arg(0)
		if romFile == "" {
			def := "BASIC.ROM"
			roms[0] = &def
		} else {
			roms[0] = &romFile
		}

	}

	env := newEnvironment(*traceCPU,
		(*traceMOS) || (*traceMOSFull),
		*traceMOSFull,
		*traceMemory,
		*panicOnErr,
		*rawline)
	defer env.close()
	handleControlC(env)

	env.mem.loadFirmware()

	for i, rom := range roms {
		if *rom != "" {
			env.mem.loadRom(*rom, uint8(0xf-i))
		}
	}

	env.initUpperLanguage()
	env.mem.completeWithRam()

	RunMOS(env)
}

func handleControlC(env *environment) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\r- Ctrl+C pressed in Terminal")
		env.close()
		os.Exit(0)
	}()
}
