package main

import "fmt"

type vdu struct {
	queue []uint8

	mode uint8

	// Mode 0-6
	textColour  uint8
	graphColour uint8

	// Mode 7
	m7fgColour uint8
	m7bgColour uint8
	m7Flash    bool

	// Toogles
	printer  bool // VDU2 and VDU3
	textOnGr bool // VDU5 and VDU4
	ignore   bool // VDU21 and VDU6
	paged    bool // VDU14 AND VDU15
}

var argsNeeded [256]int

func newVdu() *vdu {
	// Init args needed array, 0 for all except:
	argsNeeded[1] = 1
	argsNeeded[17] = 1
	argsNeeded[18] = 2
	argsNeeded[19] = 3
	argsNeeded[20] = 4
	argsNeeded[22] = 1
	argsNeeded[23] = 9
	argsNeeded[24] = 8
	argsNeeded[25] = 5
	argsNeeded[28] = 4
	argsNeeded[29] = 4
	argsNeeded[31] = 2

	var v vdu
	// Mode 7 on startup
	v.mode = 7
	v.m7fgColour = 7 // white

	return &v
}

func (v *vdu) writeAscii(i uint8) {
	if i == 0x0d {
		v.writeNewline()
	} else {
		v.write(i)
	}
}

func (v *vdu) writeNewline() {
	v.write('\r')
	v.write('\n')
}

func (v *vdu) clearQueue() {
	v.queue = nil
}

func (v *vdu) write(i uint8) {
	if v.queue == nil {
		if argsNeeded[i] == 0 {
			// Single byte command
			v.writeInternal(i, nil)
		} else {
			// Store
			v.queue = []uint8{i}
		}
	} else {
		// Store
		v.queue = append(v.queue, i)
		if argsNeeded[v.queue[0]] == len(v.queue)-1 {
			// We have enough args
			v.writeInternal(v.queue[0], v.queue[1:])
			v.queue = nil
		}
	}
}

func (v *vdu) writeInternal(cmd uint8, q []uint8) {
	out := ""
	switch cmd {
	case 0:
		// This code is ignored.
	case 1:
		/*
			This code causes the next character to be sent to the printer only and not to the
			screen. The printer must already have been enabled with VDU2. Many printers
			use special control characters to change, for example, the size of the printed
			output. For example the Epson FX-80 requires a code 14 to place it into double
			width print mode. This could be effected with the statement
				VDU1,14
			or by pressing CTRL A and then CTRL N. This code also enables the ‘printer
			ignore’ character selected by *FX6 to be sent to the printer.
		*/
		// TODO: write to printer
	case 2:
		/*
		   This code turns the printer on which means that all output to the screen will
		   also be sent to the printer. In a program the statement VDU2 should be used, but
		   the same effect can be obtained by typing CTRL B.
		*/
		v.printer = true
	case 3:
		/*
		   This code turns the printer off. No further output will be sent to the printer
		   after the statement VDU3 or after typing CTRL C
		*/
		v.printer = false
	case 4:
		/*
		   This code causes text to be written at the text cursor, ie in the normal fashion.
		   A MODE change selects VDU4, normal operation.
		*/
		v.textOnGr = false
	case 5:
		/*
		   This code causes text to be written where the graphics cursor is. The position
		   of the text cursor is unaffected. Normally the text cursor is controlled with
		   statements such as
		   		PRINT TAB(5,10);
		   and the graphics cursor is controlled with statements like
		   		MOVE700,450
		   Once the statement VDU5 has been given only one cursor is active (the graphics
		   cursor). This enables text characters to be placed at any position on the screen.
		   There are a number of other effects: text characters overwrite what is already on
		   the screen so that characters can be superimposed; text and graphics can only be
		   written in the graphics window and the colours used for both text and graphics
		   are the graphics colours. In addition the page no longer scrolls up when at the
		   bottom of the page. Note however that POS and VPOS still give you the position
		   of the text cursor.
		*/
		v.textOnGr = true
	case 6:
		/*
		   VDU6 is a complementary code to VDU21. VDU21 stops any further characters
		   being printed on the screen and VDU6 re-enables screen output. A typical use for
		   this facility would be to prevent a password appearing on the screen as it is being
		   typed in.
		*/
		v.ignore = false
	case 7:
		/*
		   This code, which can be entered in a program as VDU7 or directly from the
		   keyboard as CTRL G, causes the computer to make a short ‘beep’. This code is
		   not normally passed to the printer.
		*/
		out = string(cmd)
	case 8:
		/*
		   8 This code (VDU8 or CTRL H) moves the text cursor one space to the left. If the
		   cursor was at the start of a line then it will be moved to the end of the previous
		   line. It does not delete characters – unlike VDU127.
		*/
		out = string(cmd)
	case 9:
		/*
		   9 This code (VDU9 or CTRL I or TAB) moves the cursor forward one character
		   position.
		*/
		out = "\x1b[D"
	case 10:
		/*
		   This statement (VDU10 or CTRL J) will move the cursor down one line. If the
		   cursor is already on the bottom line then the whole display will normally be
		   moved up one line.
		*/
		//out = "\x1b[B" // Not the normal \n, here we need \r\n
		out = v.mode7ResetCode() + string(cmd)
	case 11:
		/*
		   This code (VDU11 or CTRL K) moves the text cursor up one line. If the cursor
		   is at the top of the screen then the whole display will move down a line.
		*/
		out = "\x1b[A"
	case 12:
		/*
		   This code clears the screen – or at least the text area of the screen. The screen
		   is cleared to the text background colour which is normally black. The BASIC
		   statement CLS has exactly the same effect as VDU12, or CTRL L. This code also
		   moves the text cursor to the top left of the text window.
		*/
		out = v.mode7ResetCode() + "\x1b[2J\x1b[H"
	case 13:
		/*
		   This code is produced by the RETURN key. However, its effect on the screen
		   display if issued as a VDU13 or PRINT CHR$(13);is to move the text cursor to
		   the left hand edge of the current text line (but within the current text window, of
		   course).
		*/
		out = v.mode7ResetCode() + string(cmd)
	case 14:
		/*
		   This code makes the screen display wait at the bottom of each page. It is
		   mainly used when listing long programs to prevent the listing going past so fast
		   that it is impossible to read. The computer will wait until a SHIFT key is pressed
		   before continuing. This mode is called ‘paged mode’. Paged mode is turned on
		   with CTRL N and off with CTRL O. When the computer is waiting at the bottom
		   of a page both the shift lock and caps lock lights will be illuminated.
		*/
		v.paged = true
	case 15:
		/*
		   This code causes the computer to leave paged mode.
		*/
		v.paged = false
	case 16:
		/*
		   This code (VDU16 or CTRL P) clears the graphics area of the screen to the
		   graphics background colour and the BASIC statement CLG has exactly the same
		   effect. The graphics background colour starts off as black but may have been
		   changed with the GCOL statement. VDU16 does not move the graphics cursor – it
		   just clears the graphics area of the screen.
		*/
		// TODO: graphics
	case 17:
		/*
		   VDU17 is used to change the text foreground and background colours. In
		   BASIC the statement COLOUR is used for an identical purpose. VDU17 is followed
		   by one number which determines the new colour. See the BASIC keyword
		   COLOUR for more details.
		*/
		v.textColour = q[0]
	case 18:
		/*
			This code allows the definition of the graphics foreground and background
			colours. It also specifies how the colour is to be placed on the screen. The colour
			can be plotted directly, ANDed, ORed or Exclusive-ORed with the colour already
			there, or the colour there can be inverted. In BASIC this is called GCOL.
			The first byte specifies the mode of action as follows:
				0 Plot the colour specified.
				1 OR the specified colour with that already there.
				2 AND the specified colour with that already there.
				3 Exclusive-OR the specified colour with that already there.
				4 Invert the colour already there.
			The second byte defines the logical colour to be used in future. If the byte is
			greater than 127 then it defines the graphics background colour (modulo the
			number of colours available). If the byte is less than 128 then it defines the
			graphics foreground colour (modulo the number of colours available).
		*/
		v.graphColour = q[0]
	case 19:
		/*
			This code is used to select the actual colour that is to be displayed for each
			logical colour. The statements COLOUR (and GCOL) are used to select the logical
			colour that is to be text (and graphics) in the immediate future. However the
			actual colour can be redefined with VDU19. For example
				MODE 5
				COLOUR 1
			will print all text in colour 1 which is red by default. However the addition of
				VDU 19,1,4,0,0,0 or VDU 19,1,4;0;
			will set logical colour 1 to actual colour 4 (blue). The three zeros after the actual
			colour in the VDU19 statement are for future expansion.
			In MODE 5 there are four colours (0, 1, 2 and 3). An attempt to set colour 4 will in
			fact set colour 0 so the statement
				VDU 19,4,4,0,0,0 or VDU 19,4,4;0;
			is equivalent to
				VDU 19,0,4,0,0,0 or VDU 19,0,4;0;
			We say that logical colours are reduced modulo the number of colours available in
			any particular MODE.
		*/
		// TODO
	case 20:
		/*
			This code (VDU20 or CTRL T) resets text and graphics foreground logical
			colours to their default values and also programs default logical to actual colour
			relationships. The default values are:
				Two colour MODEs
					0=Black
					1=White
				Four colour MODEs
					0=Black
					1=Red
					2=Yellow
					3=White
				16 colour MODE
					0=Black
					1=Red
					2=Green
					3=Yellow
					4=Blue
					5=Magenta
					6=Cyan
					7=White
					8=Flashing black/white
					9=Flashing red/eyan
					10=Flashing green/magenta
					11=Flashing yellow/blue
					12=Flashing blue/yellow
					13=Flashing magenta/green
					14=Flashing cyan/red
					15=Flashing white/black353
		*/
		// TODO: non mode 7 colours
	case 21:
		/*
			This code behaves in two different ways. If entered at the keyboard (as CTRL
			U) it can be used to delete the whole of the current line. It is used instead of
			pressing the DELETE key many times. If the code is generated from within a
			program by either VDU21 or PRINT CHR$(21); it has the effect of stopping all
			further graphics or text output to the screen. The VDU is said to be disabled. It
			can be enabled with VDU6.
		*/
		v.ignore = true
	case 22:
		/*
			This VDU code is used to change MODE. It is followed by one number which is
			the new MODE. Thus VDU22,7 is exactly equivalent to MODE 7 (except that it
			does not change HIMEM).
		*/
		v.mode = q[0]
		out = v.mode7ResetCode()
	case 23:
		/*
			This code is used to reprogram displayed characters. The ASCII code assigns
			code numbers for each displayed letter and number. The normal range of
			displayed characters includes all upper and lower case letters, numbers and
			punctuation marks as well as some special symbols. These characters occupy
			ASCII codes 32 to 126. If the user wishes to define his or her own characters or
			shapes then ASCII codes 224 to 255 are left available for this purpose. In fact you
			can redefine any character that is displayed, but extra memory must be set aside
			if this is done.
		*/
		// TODO: user defined characters
	case 24:
		/*
		   This code enables the user to define the graphics window – that is, the area of
		   the screen inside which graphics can be drawn with the DRAW and PLOT statements.
		   [...] When defining a graphics window four coordinates must be given; the left,
		   bottom, right and top edges of the graphics area. Suppose that we wish to confine
		   all graphics to the area shown below. [...]
		   For those who wish to know why trailing semi-colons are used the reason is as
		   follows: X and Y graphics coordinates have to be sent to the VDU software as two
		   bytes since the values may well be greater than 255. The semi-colon punctuation
		   in the VDU statement sends the number as a two byte pair with low byte first
		   followed by the high byte.
		*/
		// TODO: graphic window
	case 25:
		/*
		   This VDU code is identical to the BASIC PLOT statement. Only those writing
		   machine code graphics will need to use it. VDU25 is followed by five bytes. The
		   first gives the value of K referred to in the explanation of PLOT in the BASIC
		   keywords chapter. The next two bytes give the X coordinate and the last two
		   bytes give the Y coordinate. Refer to the entry for VDU24 for an explanation of
		   the semi-colon syntax used.
		   For example
		   		VDU 25,4,100;500;
		   would move to absolute position 100,500.
		   The above is completely equivalent to
		   		VDU 25,4,100,0,244,1
		*/
		// TODO: graphic plot
	case 26:
		/*
		   The code VDU26 CTRL Z) returns both the graphics and text windows to
		   their initial values where they occupy the whole screen. This code repositions the
		   text cursor at the top left of the screen, the graphics cursor at the bottom left and
		   sets the graphics origin to the bottom left of the screen. In this state it is possible
		   to write text and to draw graphics anywhere on the screen.
		*/
		out = v.mode7ResetCode() + "\x1b[H"
		// TODO: graphics reset
	case 27:
		/*
		   This code does nothing.
		*/
	case 28:
		/*
			   This code (VDU28) is used to set a text window. Initially it is possible to write
			   text anywhere on the screen but establishing a text window enables the user to
			   restrict all future text to a specific area of the screen. The format of the
			   statement is
			   		VDU 28,leftX,bottomY,rightX,topY
			   where
					leftX sets the left hand edge of the window
					bottomY sets the bottom edge
					rightX sets the right hand edge
					topY sets the top edge
			   [...]
			   Note that the units are character positions and the maximum values will depend
			   on the MODE in use.
		*/
		// TODO: text window
	case 29:
		/*
		   This code is used to move the graphics origin. The statement VDU29 is
		   followed by two numbers giving the X and Y coordinates of the new origin. The
		   graphics screen is addressed as shown below:
		   To move the origin to the centre of the screen the statement
		   		VDU 29,640;400;
		   should be executed. Note that the X and Y values should be followed by semi-
		   colons. See the entry for VDU24 if you require an explanation of the trailing
		   semi-colons. Note also that the graphics cursor is not affected by VDU29.
		*/
		// TODO: graphics origin
	case 30:
		/*
		   This code (VDU30 or CTRL ^) moves the text cursor to the top left of the text
		   area.
		*/
		out = v.mode7ResetCode() + "\x1b[H"
	case 31:
		/*
		   The code VDU31 enables the text cursor to be moved to any character position
		   on the screen. The statement VDU31 is followed by two numbers which give the
		   X and Y coordinates of the desired position.
		   To move the text cursor to the centre of the screen in MODE 7 one would execute
		   the statement
		   		VDU 31,20,10
		   Note that the maximum values of X and Y depend on the MODE selected and that
		   both X and Y are measured from the edges of the current text window not the
		   edges of the screen.
		*/
		// TODO: move text cursor
	case 127:
		/*
		   127 This code moves the text cursor back one character and deletes the
		   character at that position. VDU127 has exactly the same effect as the DELETE
		   key.
		*/
		out = "\x1b[D \x1b[D"

	default:
		if v.mode == 7 {
			/*
				The order of colors is: black, red, green, yellow, blue, magenta,
				cyan and white for mmode 7 and for ANSI.
				Red is:
				 - 129 as control code
				 . 1 as m7fgColour and m7bgColour
				 . 31 as ANSI fg color
				 - 41 as ANSI bg color
			*/
			switch {
			case 129 <= cmd && cmd <= 135: // Text colors
				v.m7fgColour = cmd - 129 + 1
				out = fmt.Sprintf("\x1b[%vm ", v.m7fgColour+30)
			case cmd == 136: // Flash
				v.m7Flash = true
				out = "\x1b[5m "
			case cmd == 137: // Steady (not flash)
				v.m7Flash = false
				out = "\x1b[25m "
			case cmd == 156: // Black background
				v.m7bgColour = 0
				out = fmt.Sprintf("\x1b[%vm ", v.m7bgColour+40)
			case cmd == 157: // New background
				v.m7bgColour = v.m7fgColour
				out = fmt.Sprintf("\x1b[%vm ", v.m7bgColour+40)
			case 128 <= cmd && cmd <= 159: // Rest
				out = " "
			default:
				out = adjustAsciiMode7(cmd)
			}

		} else {
			// Modes 0 to 6
			switch {
			case 32 <= cmd && cmd <= 126:
				/*
				   32-126 These codes generate the full set of letters and numbers in the ASCII set.
				   See the ASCII codes in the Appendices.
				*/
				out = adjustAscii(cmd)
			case 128 <= cmd && cmd <= 223:
				/*
				   128-223 These characters are normally undefined and will produce random
				   shapes.
				*/
				out = " "
			case 224 <= cmd && cmd <= 255:
				/*
				   224-255 These characters may be defined by the user using the statement VDU23.
				*/
				out = " "
			default:
				// It should never happen. This switch should be exhaustive
				panic("Missing VDU code")
			}
		}

	}

	if out != "" && !v.ignore {
		fmt.Print(out)
	}
}

func (v *vdu) mode7ResetCode() string {
	if v.mode != 7 {
		return ""
	}
	out := ""
	if v.m7fgColour != 7 /* white */ {
		out += "\x1b[37m"
		v.m7fgColour = 7
	}
	if v.m7bgColour != 0 /* black */ {
		out += "\x1b[40m"
		v.m7bgColour = 0
	}
	if v.m7Flash {
		out += "\x1b[25m"
		v.m7Flash = false
	}
	return out
}

func (v *vdu) mode7Reset() {
	fmt.Print(v.mode7ResetCode())
}

func adjustAscii(ch uint8) string {
	// Some chars are different from standard ASCII
	// See: http://beebwiki.mdfs.net/ASCII
	switch ch {
	case '`':
		return "£"
	case '|':
		return "¦"
	}

	return string(ch)
}

func adjustAsciiMode7(ch uint8) string {
	// Some chars are different from standard ASCII
	// See: http://beebwiki.mdfs.net/ASCII
	ch = ch & 0x7f

	switch ch {
	case '[':
		return "←"
	case '\\':
		return "½"
	case ']':
		return "→"
	case '^':
		return "↑"
	case '_':
		return "–"
	case '`':
		return "£"
	case '{':
		return "¼"
	case '|':
		return "‖"
	case '}':
		return "¾"
	case '~':
		return "÷"
	}

	return string(ch)
}
