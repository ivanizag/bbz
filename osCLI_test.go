package main

import (
	"os"
	"strings"
	"testing"
)

func Test_OSCLI_HELP(t *testing.T) {
	out := integrationTestBasic([]string{
		"*HELP",
	})

	if !strings.Contains(out, "BBZ") {
		t.Log(out)
		t.Error("*HELP is not returning BBZ")
	}
}

func Test_OSCLI_FX_commas(t *testing.T) {
	out := integrationTestBasic([]string{
		"*FX 200,3",
		"IF ERR=40 PRINT \"PA\" + \"SS\"",
	})

	if !strings.Contains(out, "PASS") {
		t.Log(out)
		t.Error("*FX error")
	}
}

func Test_OSCLI_FX_spaces(t *testing.T) {
	out := integrationTestBasic([]string{
		"*FX200 3 1",
		"IF ERR=40 PRINT \"PA\" + \"SS\"",
	})

	if !strings.Contains(out, "PASS") {
		t.Log(out)
		t.Error("*FX error")
	}
}

func Test_OSCLI_CAT(t *testing.T) {
	source := "Hi []{}"
	printed := "Hi ←→¼¾>"

	// Prepare a file
	file, err := os.CreateTemp("", "bbztest.*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(file.Name())

	_, err = file.WriteString(source)
	if err != nil {
		t.Fatal(err)
	}

	out := integrationTestBasic([]string{
		"*CAT " + file.Name(),
	})

	if !strings.Contains(out, printed) {
		t.Log(out)
		t.Error("*CAT error")
	}

	out2 := integrationTestBasic([]string{
		"*CAT " + file.Name(),
	})

	if !strings.Contains(out2, printed) {
		t.Log(out)
		t.Error("*. error")
	}

}
