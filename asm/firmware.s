; build:
;   cl65 --target none -l asm/firmware.lst asm/firmware.s --config asm/firmware.cfg -o firmware
; Review constants.go when this code is changed


; constants
                .exportzp ROM_SELECT := $f4
                .export ROM_TABLE := $023a
                .export LANGUAGE_ENTRY := $8000
                .export SERVICE_ENTRY := $8003
                .export ROM_LATCH := $fe30


; boot code
                .org $0000
START:          nop


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

; support code
                .res $f000 - *
                .org $f000

; Send cli command to ROMS and check the result
; See https://github.com/raybellis/mos120/blob/2e2ff80708e79553643e4b77c947b0652117731b/mos120.s#L10701
; Expects A=4, X=F, Y=0, the command to be pointed by $f2
CLITOROMS:      tax                     ; Service call number
                jsr OSBYTE_143
                beq CTR_CLAIMED
                brk                     ; "254-Bad command" error
                .byte $fe
                .asciiz "Bad command"
CTR_CLAIMED:    rts

;*************************************************************************
;*
;*  OSBYTE 143: Pass service commands to sideways ROMs
        ; On entry X=service call number
        ; Y=any additional parameter
        ; On entry X=0 if claimed, or preserved if unclaimed
        ; Y=any returned parameter
        ; When called internally, EQ set if call claimed
;* See https://github.com/raybellis/mos120/blob/2e2ff80708e79553643e4b77c947b0652117731b/mos120.s#L10683

OSBYTE_143:     lda ROM_SELECT          ; Get current ROM number
                pha                     ; Save it
                txa                     ; Pass service call number to  A
                ldx #$0f                ; Start at ROM 15
                                        ; Issue service call loop
NEXT:           inc ROM_TABLE,X         ; Read bit 7 on ROM type table (no ROM has type 254 &FE)
                dec ROM_TABLE,X         ;
                bpl SKIP                ; If not set (+ve result), step to next ROM down
                stx ROM_SELECT          ; Otherwise, select this ROM, &F4 RAM copy
                stx ROM_LATCH           ; Page in selected ROM
                jsr SERVICE_ENTRY       ; Call the ROM's service entry
                                        ; X and P do not need to be preserved by the ROM
                tax                     ; On exit pass A to X to chech if claimed
                beq CLAIMED             ; If 0, service call claimed, reselect ROM and exit
                ldx ROM_SELECT          ; Otherwise, get current ROM back
SKIP:           dex                     ; Step to next ROM down
                bpl NEXT                ; Loop until done ROM 0

CLAIMED:        pla                     ; Get back original ROM number
                sta ROM_SELECT          ; Set ROM number RAM copy
                sta ROM_LATCH           ; Page in the original ROM
                txa                     ; Pass X back to A to set zero flag
                rts                     ; And return



; area to store an error message
                .res $fa00 - *
                .org $fa00
errorArea:      brk
errorCode:      .byte 0
errorMessage:   .asciiz "Hello world"

; bbz host entry points
                .res $fb00 - *
                .org $fb00
epUPT:          rts                     ; 0xfb00
epEVNT:         rts                     ; 0xfb01
epFSC:          rts                     ; 0xfb02
epFIND:         rts                     ; 0xfb03
epGBPB:         rts                     ; 0xfb04
epBPUT:         rts                     ; 0xfb05
epBGET:         rts                     ; 0xfb06
epARGS:         rts                     ; 0xfb07
epFILE:         rts                     ; 0xfb08
epRDCH:         rts                     ; 0xfb09
epWRCH:         rts                     ; 0xfb0a
epWORD:         rts                     ; 0xfb0b
epBYTE:         rts                     ; 0xfb0c
epCLI:          rts                     ; 0xfb0d
epIRQ2:         rts                     ; 0xfb0e
epIRQ1:         rts                     ; 0xfb0f
epBRK:          rts                     ; 0xfb10
epUSER:         rts                     ; 0xfb11
epSYSBRK:       rts                     ; 0xfb12
epRDRM:         rts                     ; 0xfb13
epVDUCH:        rts                     ; 0xfb14
epGSINIT:       rts                     ; 0xfb15
epGSREAD:       rts                     ; 0xfb16
epNET:          rts                     ; 0xfb17
epVDU:          rts                     ; 0xfb18
epKEY:          rts                     ; 0xfb19
epINS:          rts                     ; 0xfb1a
epREM:          rts                     ; 0xfb1b
epCNP:          rts                     ; 0xfb1c
epIND1:         rts                     ; 0xfb1d
epIND2:         rts                     ; 0xfb1e
epIND3:         rts                     ; 0xfb1f


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
OSGBGP:         jmp (GBPBV)             ; OSGBPB transfer block to or from a file
OSBPUT:         jmp (BPUTV)             ; OSBPUT save a byte to file
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
                .addr $0000             ; NMI address
                .addr LANGUAGE_ENTRY    ; RESET address
                .addr epSYSBRK          ; IRQ address
