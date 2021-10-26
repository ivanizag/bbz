package main

func osByte166to255(env *environment, a uint8, x uint8, y uint8) (uint8, uint8, string) {
	/*
		Set newValue = (oldValue AND Y) EOR X

		.osbyte166to255EntryPoint = $e99c
			TAY                                                 Y=A
			LDA .mosVariablesMinus166,Y                         read current value
			TAX                                                 preserve this in X
			AND .osbyteY                                        }
			EOR .osbyteX                                        } new value = (old value AND Y) EOR X
			STA .mosVariablesMinus166,Y                         store it
			LDA .mosVariablesMinus166+1,Y                       get value of next byte into A
			TAY                                                 Y=A
			RTS
	*/

	address := mosVariablesStart + uint16(a) - 0xa6
	oldValue := env.mem.Peek(address)
	newValue := (oldValue & y) ^ x
	updateOSVar(env, a, newValue)

	return oldValue, env.mem.Peek(address + 1), mosVariableNames[a]
}

func updateOSVar(env *environment, a uint8, value uint8) {
	address := mosVariablesStart + uint16(a) - 0xa6
	env.mem.Poke(address, value)

	if a == 0xda {
		env.vdu.clearQueue()
	}
}

var mosVariableNames [256]string

func initOSVars(env *environment) {
	f := func(a uint8, name string, value uint8) {
		updateOSVar(env, a, value)
		mosVariableNames[a] = name
	}

	f(0xa8, "adress of extended vector table LO", uint8(extentedVectorTableStart&0xff))
	f(0xa9, "adress of extended vector table HI", uint8(extentedVectorTableStart>>8))
	f(0xda, "Number of items in VDU queue", 0)
	f(0xec, "Character output device status", 0)

	/*
		This location contains a value indicating the type of the last BREAK performed.
			value 0 - soft BREAK
			value 1 - power up reset
			value 2 - hard BREAK
	*/
	f(0xfd, "Hard/soft break", 1)
}
