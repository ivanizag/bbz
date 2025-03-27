package main

import (
	"os"
	"strings"
)

func integrationTestBasic(lines []string) string {

	def := "BASIC.ROM"
	roms := []*string{&def}

	env := newEnvironment(roms, false, false, false, false, false)
	con := newConsoleMock(env, lines)
	env.con = con
	RunMOS(env)
	return con.output
}

const TEST_FILE_PLACEHOLDER = "$$file$$"

func integrationTestBasicWithFile(lines []string, fileContent string) (string, error) {

	file, err := os.CreateTemp("", "bbztest.*.txt")
	if err != nil {
		return "", err
	}
	defer os.Remove(file.Name())

	_, err = file.WriteString(fileContent)
	if err != nil {
		return "", err
	}

	for i, line := range lines {
		if strings.Contains(line, TEST_FILE_PLACEHOLDER) {
			lines[i] = strings.ReplaceAll(line, TEST_FILE_PLACEHOLDER, file.Name())
		}
	}

	def := "BASIC.ROM"
	roms := []*string{&def}

	env := newEnvironment(roms, false, false, false, false, false)
	con := newConsoleMock(env, lines)
	env.con = con
	RunMOS(env)
	return con.output, nil
}
