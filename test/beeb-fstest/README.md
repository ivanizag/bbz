# BBC Micro filing system tests

Code that exercises a number of the usual disk filing system entry
points and checks the results are... what's the right term here?
"Sensible", I suppose. I'm not promising more than that. There's not
even a right or wrong in every case, as the documentation isn't always
consistent.

The goal is to produce some tests that pass on every filing system
tested (unless any of them do anything obviously ridiculous). This can
serve partly as documentation of what one can except from the average
BBC Micro filing system, and partly as a mini test suite for any
future filing systems.

# how to run

To run, load `0/$.FSTEST` - a tokenized BBC BASIC file - on the BBC
Micro.

If the filing system is DFS-type, activate it, put blank disk in drive
0, and run the program.

If it's ADFS-type, mount blank disk, modify `ADFSDRIVE$` if the disk
isn't in drive 0, and run the program.

The test will stop with a `STOP` if anything surprising happens.

6502 second processor recommended - this won't speed up the disk
access, of course, but some of the checks are a bit slow in BASIC at
2MHz.

# ADFS/DFS differences

The tests generally do the same actions on all filing systems, but
there are some unavoidable differences between DFS and ADFS, which the
tests manage as follows:

* ADFS mode never changes drive; DFS mode selects (but doesn't access)
  drive 2

* ADFS mode never selects a directory that isn't known to exist; DFS
  mode does
  
* ADFS mode sets file attributes to 3 (WR) or 1 (R); DFS mode sets
  them to 0 (unlocked) or 8 (locked)

ADFS is detected by checking for filing system 8. Add extra cases if
necessary. PRs welcome.

# pass

Tests run as-is to completion.

## DFS 2.26

* OSFILE A=0 returns 1 in accumulator whether the file was newly
  created or not
* OSFILE A=7 not supported
* File names retrieved by OSGBPB are padded to 7 chars with spaces

## DFS 2.24 (Master MOS 3.20), DFS 2.45 (Master MOS 3.50)

* OSFILE A=0 returns 1 in accumulator whether the file was newly
  created or not
* File names retrieved by OSGBPB are padded to 7 chars with spaces
  
## ADFS 1.50 (Master MOS 3.20), ADFS 2.03 (Master MOS 3.50)

* OSFILE A=5 modifies LSB of parameter block attributes only
* OSFILE A=2/3/4 can modify the current directory if directed to
  change attributes of files in another directory - suspect this is a
  bug - haven't looked into this in any detail yet. The tests work
  around this by doing `*DIR $` after each test
* Directory and file names retrieved by OSGBPB are padded to 10 chars
  with spaces

## BeebLink

* OSFILE A=0 returns 1 in accumulator whether the file was newly
  created or not (this is arguably incorrect, but it's trying to copy
  DFS...)
* OSFILE A=7 returns 1 in accumulator

## Opus DDOS 3.45, Opus Challenger 1.03

* OSFILE A=0 returns 1 in accumulator whether the file was newly
  created or not
* OSFILE treats both bits 1 and 3 of attributes as the locked flag
* OSFILE A=2/3 can't be used to change load/exec addresses of a locked
  file - the call fails with a `File locked` BRK
* OSFILE A=7 not supported
* Directory and file names retrieved by OSGBPB are padded to 7 chars
  with spaces

## Opus DDOS 3.45

* `*DIR` doesn't support the `:DRIVE.DIR` syntax.

# partial pass

Tests run to completion with minor modifications.

## Watford DDFS 1.53, Watford DDFS 1.54T

The `OSFILE SAVE` and `OSFILE SAVE (REPLACE EXISTING)` tests can be
modified or removed to cater for the A=0 behaviour.

(Perhaps the OSFILE A=0 requirements should be relaxed when on a
B/B+? - the Master Reference Manual is pretty clear about what OSFILE
A=0 does, but none of the B-era docs mention this...)

* OSFILE A=0 doesn't update parameter block bytes 10-13, and sets
  bytes 14-17 to 0
* OSFILE A=7 not supported
* File names retrieved by OSGBPB are padded to 7 chars with spaces
* OSFILE A=0 returns 1 in accumulator whether the file was newly
  created or not

# fail

It's possible there's something wrong with the tests, but I've just
assumed these are FS bugs.

## Opus Challenger 1.01

* craps out with a `Bad track format` error when doing OSFILE A=0. Use
  1.03 instead, which passes
