package main

import (
	"strings"
	"testing"
)

func TestHighAscciTokens(t *testing.T) {
	out := integrationTestBasic([]string{
		"ñ12+34", // ñ is the token for PRINT
	})

	if !strings.Contains(out, "46") {
		t.Log(out)
		t.Error("beeb-fstest failed")
	}
}
