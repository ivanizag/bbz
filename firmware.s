; build:
;    cl65 -l firmware.lst firmware.s --config firmware.cfg


; constants
                .export ENTRY := $8000


; boot code
                .org $0000
BOOT:
                LDA #$1
                JMP ENTRY


; init ram vectors
                .res $0200 - *
                .org $0200
VECTORS_START:
USERV:          .addr epUSER
BRKV:           .addr epBRK
IRQ1V:          .addr epIRQ1
IRQ2V:          .addr epIRQ2
CLIV:           .addr epCLI
BYTEV:          .addr epBYTE
WORDV:          .addr epWORD
WRCHV:          .addr epWRCH
RDCHV:          .addr epRDCH
FILEV:          .addr epFILE
ARGSV:          .addr epARGS
BGETV:          .addr epBGET
BPUTV:          .addr epBPUT
GBPBV:          .addr epGBPB
FINDV:          .addr epFIND
FSCV:           .addr epFSC
EVNTV:          .addr epEVNT
UPTV:           .addr epUPT
NETV:           .addr epNET
VDUV:           .addr epVDU
KEYV:           .addr epKEY
INSV:           .addr epINS
REMV:           .addr epREM
CNPV:           .addr epCNP
IND1V:          .addr epIND1
IND2V:          .addr epIND2
IND3V:          .addr epIND3


; bbz host entry points
                .res $fb00 - *
                .org $fb00
epUPT:          rts       ; 0xfb00
epEVNT:         rts       ; 0xfb01
epFSC:          rts       ; 0xfb02
epFIND:         rts       ; 0xfb03
epGBPB:         rts       ; 0xfb04
epBPUT:         rts       ; 0xfb05
epBGET:         rts       ; 0xfb06
epARGS:         rts       ; 0xfb07
epFILE:         rts       ; 0xfb08
epRDCH:         rts       ; 0xfb09
epWRCH:         rts       ; 0xfb0a
epWORD:         rts       ; 0xfb0b
epBYTE:         rts       ; 0xfb0c
epCLI:          rts       ; 0xfb0d
epIRQ2:         rts       ; 0xfb0e
epIRQ1:         rts       ; 0xfb0f
epBRK:          rts       ; 0xfb10
epUSER:         rts       ; 0xfb11
epSYSBRK:       rts       ; 0xfb12
epRDRM:         rts       ; 0xfb13
epVDUCH:        rts       ; 0xfb14
epGSINIT:       rts       ; 0xfb16
epGSREAD:       rts       ; 0xfb17
epNET:          rts       ; 0xfb18
epVDU:          rts       ; 0xfb19
epKEY:          rts       ; 0xfb1a
epINS:          rts       ; 0xfb1b
epREM:          rts       ; 0xfb1c
epCNP:          rts       ; 0xfb1d
epIND1:         rts       ; 0xfb1e
epIND2:         rts       ; 0xfb1f
epIND3:         rts       ; 0xfb20


; MOS function calls
                .res $ffb9 - *
                .org $ffb9
OSRDRM:         jmp epRDRM              ; OSRDRM get a byte from sideways ROM
VDUCHR:         jmp epVDUCH             ; VDUCHR VDU character output
OSEVEN:         jmp epEVNT              ; OSEVEN generate an EVENT
GSINIT:         jmp epGSINIT            ; GSINIT initialise OS string
GSREAD:         jmp epGSREAD            ; GSREAD read character from input stream
NVRDCH:         jmp epRDCH              ; NVRDCH non vectored OSRDCH
NVWRCH:         jmp epWRCH              ; NVWRCH non vectored OSWRCH
OSFIND:         jmp (FINDV)             ; OSFIND open or close a file
                jmp (GBPBV)             ; OSGBPB transfer block to or from a file
OSBPUT:	        jmp (BPUTV)             ; OSBPUT save a byte to file
OSBGET:         jmp (BGETV)             ; OSBGET get a byte from file
OSARGS:         jmp (ARGSV)             ; OSARGS read or write file arguments
OSFILE:         jmp (FILEV)             ; OSFILE read or write a file
OSRDCH:         jmp (RDCHV)             ; OSRDCH get a byte from current input stream
OSASCI:         cmp #$0d                ; OSASCI output a byte to VDU stream expanding
                bne OSWRCH              ; carriage returns (&0D) to LF/CR (&0A,&0D)
OSNEWL:         lda #$0a                ; OSNEWL output a CR/LF to VDU stream
                jsr OSWRCH              ; Outputs A followed by CR to VDU stream
                lda #$0d                ; OSWRCR output a CR to VDU stream
OSWRCH:         jmp (WRCHV)             ; OSWRCH output a character to the VDU stream
OSWORD:         jmp (WORDV)             ; OSWORD perform operation using parameter table
OSBYTE:         jmp (BYTEV)             ; OSBYTE perform operation with single bytes
OSCLI:          jmp (CLIV)              ; OSCLI pass string to command line interpreter


; 6502 vectors
                .addr $0000       ; NMI address
                .addr ENTRY    ; RESET address
                .addr epSYSBRK      ; IRQ address
