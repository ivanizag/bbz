;*************************************************************************
;*	 GSINIT	 string initialisation					 *
;*	 F2/3 points to string offset by Y				 *
;*									 *
;*	 ON EXIT							 *
;*	 Z flag set indicates null string,				 *
;*	 Y points to first non blank character				 *
;*	 A contains first non blank character				 *
;*************************************************************************


.exportzp OSBYTE_PAR_3		:= $e4
.exportzp OSBYTE_PAR_2		:= $e5
.exportzp TEXT_PTR		:= $f2

_GSINIT:		ror	OSBYTE_PAR_3			; Rotate moves carry to &E4
			jsr	_SKIP_SPACE			; get character from text
			iny					; increment Y to point at next character
			cmp	#$22				; check to see if its '"'
			beq	_BEA2A				; if so EA2A (carry set)
			dey					; decrement Y
			clc					; clear carry
_BEA2A:			ror	OSBYTE_PAR_3			; move bit 7 to bit 6 and put carry in bit 7
			cmp	#$0d				; check to see if its CR to set Z
			rts					; and return

; Skip spaces
_SKIP_SPACES_NXT:	iny					
_SKIP_SPACE:		lda	(TEXT_PTR),Y			
			cmp	#$20				
			beq	_SKIP_SPACES_NXT		
__compare_newline:	cmp	#$0d				
			rts	

;*************************************************************************
;*	 GSREAD	 string read routine					 *
;*	 F2/3 points to string offset by Y				 *
;*									 *
;*************************************************************************
_BD9B7:			.byte	$ff				; USER 6522 Bit IRQ mask		 &277
				;
_GSREAD:		lda	#$00				; A=0
_BEA31:			sta	OSBYTE_PAR_2			; store A
			lda	(TEXT_PTR),Y			; read first character
			cmp	#$0d				; is it CR
			bne	_BEA3F				; if not goto EA3F
			bit	OSBYTE_PAR_3			; if bit 7=1 no 2nd '"' found
			bmi	_LEA8F				; goto EA8F
			bpl	_BEA5A				; if not EA5A

_BEA3F:			cmp	#$20				; is less than a space?
			bcc	_LEA8F				; goto EA8F
			bne	_BEA4B				; if its not a space EA4B
			bit	OSBYTE_PAR_3			; is bit 7 of &E4 =1
			bmi	_BEA89				; if so goto EA89
			bvc	_BEA5A				; if bit 6 = 0 EA5A
_BEA4B:			cmp	#$22				; is it '"'
			bne	_BEA5F				; if not EA5F
			bit	OSBYTE_PAR_3			; if so and Bit 7 of &E4 =0 (no previous ")
			bpl	_BEA89				; then EA89
			iny					; else point at next character
			lda	(TEXT_PTR),Y			; get it
			cmp	#$22				; is it '"'
			beq	_BEA89				; if so then EA89

_BEA5A:			jsr	_SKIP_SPACE			; read a byte from text
			sec					; and return with
			rts					; carry set
								;
_BEA5F:			cmp	#$7c				; is it '|'
			bne	_BEA89				; if not EA89
			iny					; if so increase Y to point to next character
			lda	(TEXT_PTR),Y			; get it
			cmp	#$7c				; and compare it with '|' again
			beq	_BEA89				; if its '|' then EA89
			cmp	#$22				; else is it '"'
			beq	_BEA89				; if so then EA89
			cmp	#$21				; is it !
			bne	_BEA77				; if not then EA77
			iny					; increment Y again
			lda	#$80				; set bit 7
			bne	_BEA31				; loop back to EA31 to set bit 7 in next CHR
_BEA77:			cmp	#$20				; is it a space
			bcc	_LEA8F				; if less than EA8F Bad String Error
			cmp	#$3f				; is it '?'
			beq	_BEA87				; if so EA87
			jsr	_LEABF				; else modify code as if CTRL had been pressed
			bit	_BD9B7				; if bit 6 set
			bvs	_BEA8A				; then EA8A
_BEA87:			lda	#$7f				; else set bits 0 to 6 in A

_BEA89:			clv					; clear V
_BEA8A:			iny					; increment Y
			ora	OSBYTE_PAR_2			; 
			clc					; clear carry
			rts					; Return
								;
_LEA8F:			brk					; 
			.byte	$fd				; error number
			.byte	"Bad string"			; message
			brk					; 


;************ Modify code as if SHIFT pressed *****************************

_LEA9C:			cmp	#$30				; if A='0' skip routine
			beq	_BEABE				; 
			cmp	#$40				; if A='@' skip routine
			beq	_BEABE				; 
			bcc	_BEAB8				; if A<'@' then EAB8
			cmp	#$7f				; else is it DELETE

			beq	_BEABE				; if so skip routine
			bcs	_BEABC				; if greater than &7F then toggle bit 4
_BEAAC:			eor	#$30				; reverse bits 4 and 5
			cmp	#$6f				; is it &6F (previously ' _' (&5F))
			beq	_BEAB6				; goto EAB6
			cmp	#$50				; is it &50 (previously '`' (&60))
			bne	_BEAB8				; if not EAB8
_BEAB6:			eor	#$1f				; else continue to convert ` _
_BEAB8:			cmp	#$21				; compare &21 '!'
			bcc	_BEABE				; if less than return
_BEABC:			eor	#$10				; else finish conversion by toggling bit 4
_BEABE:			rts					; exit
								;
								; ASCII codes &00 &20 no change
								; 21-3F have bit 4 reverses (31-3F)
								; 41-5E A-Z have bit 5 reversed a-z
								; 5F & 60 are reversed
								; 61-7E bit 5 reversed a-z becomes A-Z
								; DELETE unchanged
								; &80+ has bit 4 changed

;************** Implement CTRL codes *************************************

_LEABF:			cmp	#$7f				; is it DEL
			beq	_BEAD1				; if so ignore routine
			bcs	_BEAAC				; if greater than &7F go to EAAC
			cmp	#$60				; if A<>'`'
			bne	_BEACB				; goto EACB
			lda	#$5f				; if A=&60, A=&5F

_BEACB:			cmp	#$40				; if A<&40
			bcc	_BEAD1				; goto EAD1  and return unchanged
			and	#$1f				; else zero bits 5 to 7
_BEAD1:			rts					; return
