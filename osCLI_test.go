package main

import (
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

func Test_OSCLI_TYPE(t *testing.T) {
	out, err := integrationTestBasicWithFile([]string{
		"*TYPE " + TEST_FILE_PLACEHOLDER,
	}, "Hi []{}")

	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(out, "Hi ←→¼¾>") {
		t.Log(out)
		t.Error("*TYPE error")
	}
}

func Test_OSCLI_EXEC(t *testing.T) {
	out, err := integrationTestBasicWithFile([]string{
		"*EXEC " + TEST_FILE_PLACEHOLDER,
	}, "PRINT 7*9")

	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(out, "63") {
		t.Log(out)
		t.Error("*EXEC error")
	}
}
