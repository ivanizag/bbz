package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/pkg/profile"
)

func main() {
	fmt.Printf("bbz - Acorn MOS for 6502 adaptation layer, https://github.com/ivanizag/bbz\n")
	fmt.Printf("(tip: uppercase is usually needed)\n")
	fmt.Printf("(press control-c twice to exit)\n\n")

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
	profileEnable := flag.Bool(
		"profile",
		false,
		"generate profile information",
	)
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

	if *profileEnable {
		// See the log with:
		//    go tool pprof --pdf ~/go/bin/izapple2sdl /tmp/profile329536248/cpu.pprof > profile.pdf
		defer profile.Start().Stop()
	}

	env := newEnvironment(roms, *traceCPU,
		(*traceMOS) || (*traceMOSFull),
		*traceMOSFull,
		*traceMemory,
		*panicOnErr)
	defer env.close()
	handleControlC(env)

	if *rawline {
		env.con = newConsoleSimple(env)
	} else {
		env.con = newConsoleLiner(env)
	}

	RunMOS(env)
}

func handleControlC(env *environment) {
	c := make(chan os.Signal, 2)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		for {
			<-c
			env.escape()
		}
	}()
}
