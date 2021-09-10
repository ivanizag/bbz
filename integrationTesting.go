package main

func integrationTestBasic(lines []string) string {

	def := "BASIC.ROM"
	roms := []*string{&def}

	env := newEnvironment(roms, false, false, false, false, false)
	con := newConsoleMock(env, lines)
	env.con = con
	RunMOS(env)
	return con.output
}
