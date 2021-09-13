package main

import (
	"os"
	"strings"
	"testing"
)

func Test_beeb_ftest(t *testing.T) {
	out := integrationTestBasic([]string{
		"LOAD \"test/beeb-fstest/0/$.FSTEST\"",
		"1661 GOTO 1690", // Skip OSGBPB 06 and 07 tests
		"1791 GOTO 2130", // Skip OSGBPB 08 tests
		"RUN",
	})

	if !strings.Contains(out, "GOOD. TOOK") {
		t.Log(out)
		t.Error("beeb-fstest failed")
	}

	os.Remove("+.0")
	os.Remove("+.0" + metadataExtension)
}
