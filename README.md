# bbz - Acorn MOS for 6502 adaptation layer

Used to run 6502 programs using the Acorn MOS on a modern operating system
as a console application. It runs BBC Micro language ROMs.

## What is this

This is a MOS for 6502 execution environment. It's a 6502 emulator that
intercepts the calls to the Acorm Machine Operating System (MOS) and
services them as a native console application on the host.

This is not a BBC Micro emulator. It does not emulate the BBC Micro
hardware, just the 6502 processor.

This program is heavily inspired on [Applecorn](https://github.com/bobbimanners/Applecorn), "a ProDOS application for the
Apple //e Enhanced which provides an environment for Acorn BBC Microcomputer
language ROMs to run."

References:
- BBC Microcomputer System User Guide: https://archive.org/details/BBCUG
- The Advanced User Guide for the BBC Microcomputer
- Applecorn source code: https://github.com/bobbimanners/Applecorn

## Features
Very few of the MOS entrypoints are defined, the minimum needed to start BBC Basic.
The program quits on any API call not supported.


## Usage 

```
bbz [flags] [filename]
```

This first arguments is the filename of the ROM to run. With no arguments it
runs the BBC Basic ROM in `BASIC.ROM`.

Avaliable flags (to put before the ROM filename if present):

``` 
  -M	dump to the console the MOS calls including console I/O calls
  -c	dump to the console the CPU execution operations
  -m	dump to the console the MOS calls excluding console I/O calls
```

## Usage example

Running BBC Basic:
```
$ ./bbz
bbz - Acorn MOS for 6502 adaptation layer, https://github.com/ivanizag/bbz

>PRINT "HELLO"
HELLO
>10 A=12
>20 PRINT A
>LIST
   10 A=12
   20 PRINT A
>RUN
        12
>HEY

Mistake
>^C
$
```

Log of the MOS calls (excluding the most verbose output API calls):
```
$ ./bbz -m
bbz - Acorn MOS for 6502 adaptation layer, https://github.com/ivanizag/bbz

[[[OSBYTE84('Read top of user mem',X=0x00,Y=0x00)]]]
[[[OSBYTE83('Read bottom of user mem',X=0x00,Y=0x80)]]]
>PRINT "HELLO"
[[[OSWORD00('read line',BUF=0x0700)='PRINT "HELLO"']]]
HELLO
>HEY
[[[OSWORD00('read line',BUF=0x0700)='HEY']]]
[[[BREAK(ERR=04, 'Mistake']]]
[[[OSBYTEda('R/W number of items in VDU',X=0x00,Y=0x00)]]]
[[[OSBYTE7e('Ack detection of an ESC condition',X=0x00,Y=0x00)]]]

Mistake
>^C
$
```

