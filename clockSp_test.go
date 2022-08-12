package main

import (
	"testing"
)

func Benchmark_ClockSp(b *testing.B) {
	for n := 0; n < b.N; n++ {
		integrationTestBasic([]string{
			"LOAD \"test/ClockSp_2.10.bas\"",
			"RUN",
		})
	}
}
