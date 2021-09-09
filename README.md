# bbz - Acorn MOS for 6502 adaptation layer

Used to run 6502 programs using the Acorn MOS on a modern operating system
as a console application. It runs BBC Micro language ROMs.

## What is this

This is a MOS for 6502 execution environment. It's a 6502 emulator that
intercepts the calls to the Acorm Machine Operating System (MOS) and
services them as a native console application on the host.

This is not a BBC Micro emulator. It does not emulate the BBC Micro
hardware, just the 6502 processor.

BBZ is a console program, it tries to look and feel as a modern console
application not as a BBC Micro. There is command history accesible with
the up arrow and control-R. Control-C behaves as the BBC Micro Escape key
to interrupt long running programs. Control-C twice will exic BBZ back to
the host.

This program is heavily inspired on [Applecorn](https://github.com/bobbimanners/Applecorn),
"a ProDOS application for the Apple //e Enhanced which provides an environment
for Acorn BBC Microcomputer language ROMs to run."

References:
- [BBC Microcomputer System User Guide](https://archive.org/details/BBCUG)
- The Advanced User Guide for the BBC Microcomputer
- [Applecorn source code](https://github.com/bobbimanners/Applecorn)
- [The MOS Reassembly for the BBC Micro](https://tobylobster.github.io/mos/mos/index.html)
- [beeb-fstest test suite](https://github.com/tom-seddon/beeb-fstest)

## Features
- Can run BBC BASIC and most of the language ROMs.
- Saves and loads files from the host filesystem.
- Readline like input with persistent history.
- Can load up to 16 sideways ROMs, the unused slots are filled with sideways RAM 16K expansions.
- Most of the MOS entrypoints and VDU control codes are defined.
- Does some of the mode 7 text coloring using ANSI escape codes on the terminal. Try `VDU 65,129,66,130,67,132,68,135,69,13,10` on BBC BASIC.
- OSCLI comands suported:
  - *|
  - */
  - *FX
  - *BASIC
  - *DIR
  - *HELP
  - *HOST cmd: execute a command on the host OS. Example: `*HOST ls -la`
  - *INFO
  - *LOAD
  - *QUIT: exit to host
  - *RUN
  - *ROMS

## Usage 

```
bbz [flags] [filename]
```

`filename` is the filename of the ROM to run (the same as `-rom0`). By default, it
uses the BBC Basic ROM in `BASIC.ROM`.

AvaIlable flags (to put before the ROM filename if present):

``` 
  -M	dump to the console the MOS calls including console I/O calls
  -c	dump to the console the CPU execution operations
  -m	dump to the console the MOS calls excluding console I/O calls
  -p	panic on not implemented MOS calls
  -r	disable readline like input with history
  -s	dump to the console the accesses to Fred, Jim or Sheila
  -rom0 string
    	filename for rom 0 (slot 0xf)
  -rom1 string
    	filename for rom 1 (slot 0xe)
  -rom2 string
    	filename for rom 2 (slot 0xd)
  -rom3 string
    	filename for rom 3 (slot 0xc)
  -rom4 string
    	filename for rom 4 (slot 0xb)
  -rom5 string
    	filename for rom 5 (slot 0xa)
  -rom6 string
    	filename for rom 6 (slot 0x9)
  -rom7 string
    	filename for rom 7 (slot 0x8)
  -rom8 string
    	filename for rom 8 (slot 0x7)
  -rom9 string
    	filename for rom 9 (slot 0x6)
  -rom10 string
    	filename for rom 10 (slot 0x5)
  -rom11 string
    	filename for rom 11 (slot 0x4)
  -rom12 string
    	filename for rom 12 (slot 0x3)
  -rom13 string
    	filename for rom 13 (slot 0x2)
  -rom14 string
    	filename for rom 14 (slot 0x1)
  -rom15 string
    	filename for rom 15 (slot 0x0)



```

## Usage example

Running BBC Basic:
```
$ ./bbz
bbz - Acorn MOS for 6502 adaptation layer, https://github.com/ivanizag/bbz

BASIC

>PRINT "HELLO"
HELLO
>10 PRINT "HEY"
>RUN
HEY
>SAVE "TEST"
>NEW
>LOAD "TEST"
>LIST
   10 PRINT "HEY"
>X

Mistake
>*HOST ls -l TEST
-rw-r--r-- 1 casa casa 14 jul 30 20:04 TEST

>^Csignal: interrupt
```

Log of the MOS calls (excluding the most verbose output API calls):
```
$ ./bbz -m ROMs/Forth_103.rom
bbz - Acorn MOS for 6502 adaptation layer, https://github.com/ivanizag/bbz

FORTH

COLD or WARM start (C/W)? C
[[[OSRDCH()=0x43]]]
[[[OSARGS('Get filing system',A=00,Y=00) => 105]]]
[[[OSBYTE82('Read machine high order address',X=0x58,Y=0x00) => (X=0xff,Y=0xff)]]]
[[[OSBYTE84('Read top of user mem',X=0x58,Y=0x00) => (X=0x00,Y=0x80)]]]
[[[OSBYTE83('Read bottom of user mem',X=0x58,Y=0x00) => (X=0x00,Y=0x0e)]]]


FORTH
OK
2 1 + .
[[[OSWORD00('read line',BUF=0x0542,range=20-ff, maxlen=80)='2 1 + .']]]
3 OK

```

Using mode 7 colors:

![mode 7 colors](doc/vdu_colors.png)

Using several ROMs at once:
```
$ ./bbz -rom0 BASIC.ROM -rom1 ROMs/Forth_103.rom -rom2 ROMs/LISP501.ROM -rom3 ROMs/COMAL.rom -rom4 ROMs/MPROLOG310.rom 
bbz - Acorn MOS for 6502 adaptation layer, https://github.com/ivanizag/bbz
(tip: uppercase is usually needed)

BASIC
>*ROMS
ROM F BASIC 01 (L)
ROM E FORTH 01 (SL)
ROM D LISP 05 (SL)
ROM C COMAL 16 (SL)
ROM B micro PROLOG  80 (SL)
RAM A 16K
RAM 9 16K
RAM 8 16K
RAM 7 16K
RAM 6 16K
RAM 5 16K
RAM 4 16K
RAM 3 16K
RAM 2 16K
RAM 1 16K
RAM 0 16K
>*HELP

BBZ 0.0

FORTH 1.03

LISP 5.01

COMAL

micro PROLOG 3.1
>*LISP
LISP


Evaluate : (* '*FORTH)
FORTH

COLD or WARM start (C/W)? C


FORTH
OK
OS' *COMAL'
COMAL
→*BASIC
BASIC
>*PROLOGUE
micro PROLOG
29184 bytes free
&.


```