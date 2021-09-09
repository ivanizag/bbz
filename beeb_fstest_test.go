package main

import (
	"strings"
	"testing"
)

func Test_beeb_ftest(t *testing.T) {

	def := "BASIC.ROM"
	roms := []*string{&def}

	env := newEnvironment(roms, false, false, false, false, false)
	con := newConsoleMock(env, []string{
		"LOAD \"test/beeb-fstest/0/$.FSTEST\"",
		"1661 GOTO 1690", // Skip OSGBPB 06 and 07 tests
		"1791 GOTO 2130", // Skip OSGBPB 08 tests
		"RUN",
	})
	env.con = con

	RunMOS(env)

	if !strings.Contains(con.output, "GOOD. TOOK") {
		t.Log(con.output)
		t.Error("beeb-fstest failed")
	}
}
