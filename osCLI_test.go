package main

import (
	"strings"
	"testing"
)

func Test_OSCLI_HELP(t *testing.T) {

	def := "BASIC.ROM"
	roms := []*string{&def}

	env := newEnvironment(roms, false, false, false, false, false)
	con := newConsoleMock(env, []string{
		"*HELP",
	})
	env.con = con

	RunMOS(env)

	if !strings.Contains(con.output, "BBZ") {
		t.Log(con.output)
		t.Error("*HELP is not returning BBZ")
	}
}
