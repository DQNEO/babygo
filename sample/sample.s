"".min STEXT nosplit size=1 args=0x0 locals=0x0
	0x0000 00000 (sample/sample.go:5)	TEXT	"".min(SB), NOSPLIT|ABIInternal, $0-0
	0x0000 00000 (sample/sample.go:5)	PCDATA	$0, $-2
	0x0000 00000 (sample/sample.go:5)	PCDATA	$1, $-2
	0x0000 00000 (sample/sample.go:5)	FUNCDATA	$0, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0000 00000 (sample/sample.go:5)	FUNCDATA	$1, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0000 00000 (sample/sample.go:5)	FUNCDATA	$2, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0000 00000 (sample/sample.go:6)	PCDATA	$0, $-1
	0x0000 00000 (sample/sample.go:6)	PCDATA	$1, $-1
	0x0000 00000 (sample/sample.go:6)	RET
	0x0000 c3                                               .
"".char STEXT nosplit size=48 args=0x10 locals=0x10
	0x0000 00000 (sample/sample.go:8)	TEXT	"".char(SB), NOSPLIT|ABIInternal, $16-16
	0x0000 00000 (sample/sample.go:8)	SUBQ	$16, SP
	0x0004 00004 (sample/sample.go:8)	MOVQ	BP, 8(SP)
	0x0009 00009 (sample/sample.go:8)	LEAQ	8(SP), BP
	0x000e 00014 (sample/sample.go:8)	PCDATA	$0, $-2
	0x000e 00014 (sample/sample.go:8)	PCDATA	$1, $-2
	0x000e 00014 (sample/sample.go:8)	FUNCDATA	$0, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x000e 00014 (sample/sample.go:8)	FUNCDATA	$1, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x000e 00014 (sample/sample.go:8)	FUNCDATA	$2, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x000e 00014 (sample/sample.go:8)	PCDATA	$0, $0
	0x000e 00014 (sample/sample.go:8)	PCDATA	$1, $0
	0x000e 00014 (sample/sample.go:8)	MOVB	$0, "".~r1+32(SP)
	0x0013 00019 (sample/sample.go:9)	MOVBLZX	"".a+24(SP), AX
	0x0018 00024 (sample/sample.go:9)	MOVB	AL, "".b+7(SP)
	0x001c 00028 (sample/sample.go:10)	MOVB	$65, "".b+7(SP)
	0x0021 00033 (sample/sample.go:11)	MOVB	$65, "".~r1+32(SP)
	0x0026 00038 (sample/sample.go:11)	MOVQ	8(SP), BP
	0x002b 00043 (sample/sample.go:11)	ADDQ	$16, SP
	0x002f 00047 (sample/sample.go:11)	RET
	0x0000 48 83 ec 10 48 89 6c 24 08 48 8d 6c 24 08 c6 44  H...H.l$.H.l$..D
	0x0010 24 20 00 0f b6 44 24 18 88 44 24 07 c6 44 24 07  $ ...D$..D$..D$.
	0x0020 41 c6 44 24 20 41 48 8b 6c 24 08 48 83 c4 10 c3  A.D$ AH.l$.H....
"".slice STEXT nosplit size=85 args=0x30 locals=0x20
	0x0000 00000 (sample/sample.go:14)	TEXT	"".slice(SB), NOSPLIT|ABIInternal, $32-48
	0x0000 00000 (sample/sample.go:14)	SUBQ	$32, SP
	0x0004 00004 (sample/sample.go:14)	MOVQ	BP, 24(SP)
	0x0009 00009 (sample/sample.go:14)	LEAQ	24(SP), BP
	0x000e 00014 (sample/sample.go:14)	PCDATA	$0, $-2
	0x000e 00014 (sample/sample.go:14)	PCDATA	$1, $-2
	0x000e 00014 (sample/sample.go:14)	FUNCDATA	$0, gclocals·305be4e7b0a18f2d72ac308e38ac4e7b(SB)
	0x000e 00014 (sample/sample.go:14)	FUNCDATA	$1, gclocals·cadea2e49003779a155f5f8fb1f0fe78(SB)
	0x000e 00014 (sample/sample.go:14)	FUNCDATA	$2, gclocals·9fb7f0986f647f17cb53dda1484e0f7a(SB)
	0x000e 00014 (sample/sample.go:14)	PCDATA	$0, $0
	0x000e 00014 (sample/sample.go:14)	PCDATA	$1, $0
	0x000e 00014 (sample/sample.go:14)	MOVQ	$0, "".~r1+64(SP)
	0x0017 00023 (sample/sample.go:14)	XORPS	X0, X0
	0x001a 00026 (sample/sample.go:14)	MOVUPS	X0, "".~r1+72(SP)
	0x001f 00031 (sample/sample.go:15)	PCDATA	$0, $1
	0x001f 00031 (sample/sample.go:15)	MOVQ	"".a+40(SP), AX
	0x0024 00036 (sample/sample.go:15)	MOVQ	"".a+48(SP), CX
	0x0029 00041 (sample/sample.go:15)	PCDATA	$1, $1
	0x0029 00041 (sample/sample.go:15)	MOVQ	"".a+56(SP), DX
	0x002e 00046 (sample/sample.go:15)	MOVQ	AX, "".b(SP)
	0x0032 00050 (sample/sample.go:15)	MOVQ	CX, "".b+8(SP)
	0x0037 00055 (sample/sample.go:15)	MOVQ	DX, "".b+16(SP)
	0x003c 00060 (sample/sample.go:16)	PCDATA	$0, $0
	0x003c 00060 (sample/sample.go:16)	PCDATA	$1, $2
	0x003c 00060 (sample/sample.go:16)	MOVQ	AX, "".~r1+64(SP)
	0x0041 00065 (sample/sample.go:16)	MOVQ	CX, "".~r1+72(SP)
	0x0046 00070 (sample/sample.go:16)	MOVQ	DX, "".~r1+80(SP)
	0x004b 00075 (sample/sample.go:16)	MOVQ	24(SP), BP
	0x0050 00080 (sample/sample.go:16)	ADDQ	$32, SP
	0x0054 00084 (sample/sample.go:16)	RET
	0x0000 48 83 ec 20 48 89 6c 24 18 48 8d 6c 24 18 48 c7  H.. H.l$.H.l$.H.
	0x0010 44 24 40 00 00 00 00 0f 57 c0 0f 11 44 24 48 48  D$@.....W...D$HH
	0x0020 8b 44 24 28 48 8b 4c 24 30 48 8b 54 24 38 48 89  .D$(H.L$0H.T$8H.
	0x0030 04 24 48 89 4c 24 08 48 89 54 24 10 48 89 44 24  .$H.L$.H.T$.H.D$
	0x0040 40 48 89 4c 24 48 48 89 54 24 50 48 8b 6c 24 18  @H.L$HH.T$PH.l$.
	0x0050 48 83 c4 20 c3                                   H.. .
"".arg1 STEXT nosplit size=1 args=0x8 locals=0x0
	0x0000 00000 (sample/sample.go:19)	TEXT	"".arg1(SB), NOSPLIT|ABIInternal, $0-8
	0x0000 00000 (sample/sample.go:19)	PCDATA	$0, $-2
	0x0000 00000 (sample/sample.go:19)	PCDATA	$1, $-2
	0x0000 00000 (sample/sample.go:19)	FUNCDATA	$0, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0000 00000 (sample/sample.go:19)	FUNCDATA	$1, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0000 00000 (sample/sample.go:19)	FUNCDATA	$2, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0000 00000 (sample/sample.go:21)	PCDATA	$0, $-1
	0x0000 00000 (sample/sample.go:21)	PCDATA	$1, $-1
	0x0000 00000 (sample/sample.go:21)	RET
	0x0000 c3                                               .
"".arg1ret1 STEXT nosplit size=20 args=0x10 locals=0x0
	0x0000 00000 (sample/sample.go:23)	TEXT	"".arg1ret1(SB), NOSPLIT|ABIInternal, $0-16
	0x0000 00000 (sample/sample.go:23)	PCDATA	$0, $-2
	0x0000 00000 (sample/sample.go:23)	PCDATA	$1, $-2
	0x0000 00000 (sample/sample.go:23)	FUNCDATA	$0, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0000 00000 (sample/sample.go:23)	FUNCDATA	$1, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0000 00000 (sample/sample.go:23)	FUNCDATA	$2, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0000 00000 (sample/sample.go:23)	PCDATA	$0, $0
	0x0000 00000 (sample/sample.go:23)	PCDATA	$1, $0
	0x0000 00000 (sample/sample.go:23)	MOVQ	$0, "".~r1+16(SP)
	0x0009 00009 (sample/sample.go:24)	MOVQ	"".x+8(SP), AX
	0x000e 00014 (sample/sample.go:24)	MOVQ	AX, "".~r1+16(SP)
	0x0013 00019 (sample/sample.go:24)	RET
	0x0000 48 c7 44 24 10 00 00 00 00 48 8b 44 24 08 48 89  H.D$.....H.D$.H.
	0x0010 44 24 10 c3                                      D$..
"".sumAndMul STEXT nosplit size=53 args=0x20 locals=0x0
	0x0000 00000 (sample/sample.go:27)	TEXT	"".sumAndMul(SB), NOSPLIT|ABIInternal, $0-32
	0x0000 00000 (sample/sample.go:27)	PCDATA	$0, $-2
	0x0000 00000 (sample/sample.go:27)	PCDATA	$1, $-2
	0x0000 00000 (sample/sample.go:27)	FUNCDATA	$0, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0000 00000 (sample/sample.go:27)	FUNCDATA	$1, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0000 00000 (sample/sample.go:27)	FUNCDATA	$2, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0000 00000 (sample/sample.go:27)	PCDATA	$0, $0
	0x0000 00000 (sample/sample.go:27)	PCDATA	$1, $0
	0x0000 00000 (sample/sample.go:27)	MOVQ	$0, "".~r2+24(SP)
	0x0009 00009 (sample/sample.go:27)	MOVQ	$0, "".~r3+32(SP)
	0x0012 00018 (sample/sample.go:28)	MOVQ	"".x+8(SP), AX
	0x0017 00023 (sample/sample.go:28)	ADDQ	"".y+16(SP), AX
	0x001c 00028 (sample/sample.go:28)	MOVQ	AX, "".~r2+24(SP)
	0x0021 00033 (sample/sample.go:28)	MOVQ	"".x+8(SP), AX
	0x0026 00038 (sample/sample.go:28)	MOVQ	"".y+16(SP), CX
	0x002b 00043 (sample/sample.go:28)	IMULQ	CX, AX
	0x002f 00047 (sample/sample.go:28)	MOVQ	AX, "".~r3+32(SP)
	0x0034 00052 (sample/sample.go:28)	RET
	0x0000 48 c7 44 24 18 00 00 00 00 48 c7 44 24 20 00 00  H.D$.....H.D$ ..
	0x0010 00 00 48 8b 44 24 08 48 03 44 24 10 48 89 44 24  ..H.D$.H.D$.H.D$
	0x0020 18 48 8b 44 24 08 48 8b 4c 24 10 48 0f af c1 48  .H.D$.H.L$.H...H
	0x0030 89 44 24 20 c3                                   .D$ .
"".sumAndMulWithNamedReturn STEXT nosplit size=53 args=0x20 locals=0x0
	0x0000 00000 (sample/sample.go:31)	TEXT	"".sumAndMulWithNamedReturn(SB), NOSPLIT|ABIInternal, $0-32
	0x0000 00000 (sample/sample.go:31)	PCDATA	$0, $-2
	0x0000 00000 (sample/sample.go:31)	PCDATA	$1, $-2
	0x0000 00000 (sample/sample.go:31)	FUNCDATA	$0, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0000 00000 (sample/sample.go:31)	FUNCDATA	$1, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0000 00000 (sample/sample.go:31)	FUNCDATA	$2, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0000 00000 (sample/sample.go:31)	PCDATA	$0, $0
	0x0000 00000 (sample/sample.go:31)	PCDATA	$1, $0
	0x0000 00000 (sample/sample.go:31)	MOVQ	$0, "".sum+24(SP)
	0x0009 00009 (sample/sample.go:31)	MOVQ	$0, "".mul+32(SP)
	0x0012 00018 (sample/sample.go:32)	MOVQ	"".x+8(SP), AX
	0x0017 00023 (sample/sample.go:32)	ADDQ	"".y+16(SP), AX
	0x001c 00028 (sample/sample.go:32)	MOVQ	AX, "".sum+24(SP)
	0x0021 00033 (sample/sample.go:33)	MOVQ	"".y+16(SP), AX
	0x0026 00038 (sample/sample.go:33)	MOVQ	"".x+8(SP), CX
	0x002b 00043 (sample/sample.go:33)	IMULQ	AX, CX
	0x002f 00047 (sample/sample.go:33)	MOVQ	CX, "".mul+32(SP)
	0x0034 00052 (sample/sample.go:34)	RET
	0x0000 48 c7 44 24 18 00 00 00 00 48 c7 44 24 20 00 00  H.D$.....H.D$ ..
	0x0010 00 00 48 8b 44 24 08 48 03 44 24 10 48 89 44 24  ..H.D$.H.D$.H.D$
	0x0020 18 48 8b 44 24 10 48 8b 4c 24 08 48 0f af c8 48  .H.D$.H.L$.H...H
	0x0030 89 4c 24 20 c3                                   .L$ .
"".sum STEXT nosplit size=25 args=0x18 locals=0x0
	0x0000 00000 (sample/sample.go:38)	TEXT	"".sum(SB), NOSPLIT|ABIInternal, $0-24
	0x0000 00000 (sample/sample.go:38)	PCDATA	$0, $-2
	0x0000 00000 (sample/sample.go:38)	PCDATA	$1, $-2
	0x0000 00000 (sample/sample.go:38)	FUNCDATA	$0, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0000 00000 (sample/sample.go:38)	FUNCDATA	$1, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0000 00000 (sample/sample.go:38)	FUNCDATA	$2, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0000 00000 (sample/sample.go:38)	PCDATA	$0, $0
	0x0000 00000 (sample/sample.go:38)	PCDATA	$1, $0
	0x0000 00000 (sample/sample.go:38)	MOVQ	$0, "".~r2+24(SP)
	0x0009 00009 (sample/sample.go:39)	MOVQ	"".x+8(SP), AX
	0x000e 00014 (sample/sample.go:39)	ADDQ	"".y+16(SP), AX
	0x0013 00019 (sample/sample.go:39)	MOVQ	AX, "".~r2+24(SP)
	0x0018 00024 (sample/sample.go:39)	RET
	0x0000 48 c7 44 24 18 00 00 00 00 48 8b 44 24 08 48 03  H.D$.....H.D$.H.
	0x0010 44 24 10 48 89 44 24 18 c3                       D$.H.D$..
"".concate STEXT size=127 args=0x30 locals=0x40
	0x0000 00000 (sample/sample.go:42)	TEXT	"".concate(SB), ABIInternal, $64-48
	0x0000 00000 (sample/sample.go:42)	MOVQ	(TLS), CX
	0x0009 00009 (sample/sample.go:42)	CMPQ	SP, 16(CX)
	0x000d 00013 (sample/sample.go:42)	PCDATA	$0, $-2
	0x000d 00013 (sample/sample.go:42)	JLS	120
	0x000f 00015 (sample/sample.go:42)	PCDATA	$0, $-1
	0x000f 00015 (sample/sample.go:42)	SUBQ	$64, SP
	0x0013 00019 (sample/sample.go:42)	MOVQ	BP, 56(SP)
	0x0018 00024 (sample/sample.go:42)	LEAQ	56(SP), BP
	0x001d 00029 (sample/sample.go:42)	PCDATA	$0, $-2
	0x001d 00029 (sample/sample.go:42)	PCDATA	$1, $-2
	0x001d 00029 (sample/sample.go:42)	FUNCDATA	$0, gclocals·2625d1fdbbaf79a2e52296235cb6527c(SB)
	0x001d 00029 (sample/sample.go:42)	FUNCDATA	$1, gclocals·f6bd6b3389b872033d462029172c8612(SB)
	0x001d 00029 (sample/sample.go:42)	FUNCDATA	$2, gclocals·1cf923758aae2e428391d1783fe59973(SB)
	0x001d 00029 (sample/sample.go:42)	PCDATA	$0, $0
	0x001d 00029 (sample/sample.go:42)	PCDATA	$1, $0
	0x001d 00029 (sample/sample.go:42)	XORPS	X0, X0
	0x0020 00032 (sample/sample.go:42)	MOVUPS	X0, "".~r2+104(SP)
	0x0025 00037 (sample/sample.go:43)	MOVQ	$0, (SP)
	0x002d 00045 (sample/sample.go:43)	PCDATA	$0, $1
	0x002d 00045 (sample/sample.go:43)	MOVQ	"".x+72(SP), AX
	0x0032 00050 (sample/sample.go:43)	PCDATA	$1, $1
	0x0032 00050 (sample/sample.go:43)	MOVQ	"".x+80(SP), CX
	0x0037 00055 (sample/sample.go:43)	PCDATA	$0, $0
	0x0037 00055 (sample/sample.go:43)	MOVQ	AX, 8(SP)
	0x003c 00060 (sample/sample.go:43)	MOVQ	CX, 16(SP)
	0x0041 00065 (sample/sample.go:43)	PCDATA	$0, $1
	0x0041 00065 (sample/sample.go:43)	MOVQ	"".y+88(SP), AX
	0x0046 00070 (sample/sample.go:43)	PCDATA	$1, $2
	0x0046 00070 (sample/sample.go:43)	MOVQ	"".y+96(SP), CX
	0x004b 00075 (sample/sample.go:43)	PCDATA	$0, $0
	0x004b 00075 (sample/sample.go:43)	MOVQ	AX, 24(SP)
	0x0050 00080 (sample/sample.go:43)	MOVQ	CX, 32(SP)
	0x0055 00085 (sample/sample.go:43)	CALL	runtime.concatstring2(SB)
	0x005a 00090 (sample/sample.go:43)	MOVQ	48(SP), AX
	0x005f 00095 (sample/sample.go:43)	PCDATA	$0, $2
	0x005f 00095 (sample/sample.go:43)	MOVQ	40(SP), CX
	0x0064 00100 (sample/sample.go:43)	PCDATA	$0, $0
	0x0064 00100 (sample/sample.go:43)	PCDATA	$1, $3
	0x0064 00100 (sample/sample.go:43)	MOVQ	CX, "".~r2+104(SP)
	0x0069 00105 (sample/sample.go:43)	MOVQ	AX, "".~r2+112(SP)
	0x006e 00110 (sample/sample.go:43)	MOVQ	56(SP), BP
	0x0073 00115 (sample/sample.go:43)	ADDQ	$64, SP
	0x0077 00119 (sample/sample.go:43)	RET
	0x0078 00120 (sample/sample.go:43)	NOP
	0x0078 00120 (sample/sample.go:42)	PCDATA	$1, $-1
	0x0078 00120 (sample/sample.go:42)	PCDATA	$0, $-2
	0x0078 00120 (sample/sample.go:42)	CALL	runtime.morestack_noctxt(SB)
	0x007d 00125 (sample/sample.go:42)	PCDATA	$0, $-1
	0x007d 00125 (sample/sample.go:42)	JMP	0
	0x0000 64 48 8b 0c 25 00 00 00 00 48 3b 61 10 76 69 48  dH..%....H;a.viH
	0x0010 83 ec 40 48 89 6c 24 38 48 8d 6c 24 38 0f 57 c0  ..@H.l$8H.l$8.W.
	0x0020 0f 11 44 24 68 48 c7 04 24 00 00 00 00 48 8b 44  ..D$hH..$....H.D
	0x0030 24 48 48 8b 4c 24 50 48 89 44 24 08 48 89 4c 24  $HH.L$PH.D$.H.L$
	0x0040 10 48 8b 44 24 58 48 8b 4c 24 60 48 89 44 24 18  .H.D$XH.L$`H.D$.
	0x0050 48 89 4c 24 20 e8 00 00 00 00 48 8b 44 24 30 48  H.L$ .....H.D$0H
	0x0060 8b 4c 24 28 48 89 4c 24 68 48 89 44 24 70 48 8b  .L$(H.L$hH.D$pH.
	0x0070 6c 24 38 48 83 c4 40 c3 e8 00 00 00 00 eb 81     l$8H..@........
	rel 5+4 t=17 TLS+0
	rel 86+4 t=8 runtime.concatstring2+0
	rel 121+4 t=8 runtime.morestack_noctxt+0
"".main STEXT size=914 args=0x0 locals=0x1e0
	0x0000 00000 (sample/sample.go:46)	TEXT	"".main(SB), ABIInternal, $480-0
	0x0000 00000 (sample/sample.go:46)	MOVQ	(TLS), CX
	0x0009 00009 (sample/sample.go:46)	LEAQ	-352(SP), AX
	0x0011 00017 (sample/sample.go:46)	CMPQ	AX, 16(CX)
	0x0015 00021 (sample/sample.go:46)	PCDATA	$0, $-2
	0x0015 00021 (sample/sample.go:46)	JLS	904
	0x001b 00027 (sample/sample.go:46)	PCDATA	$0, $-1
	0x001b 00027 (sample/sample.go:46)	SUBQ	$480, SP
	0x0022 00034 (sample/sample.go:46)	MOVQ	BP, 472(SP)
	0x002a 00042 (sample/sample.go:46)	LEAQ	472(SP), BP
	0x0032 00050 (sample/sample.go:46)	PCDATA	$0, $-2
	0x0032 00050 (sample/sample.go:46)	PCDATA	$1, $-2
	0x0032 00050 (sample/sample.go:46)	FUNCDATA	$0, gclocals·f5be5308b59e045b7c5b33ee8908cfb7(SB)
	0x0032 00050 (sample/sample.go:46)	FUNCDATA	$1, gclocals·a5f8d5878f569589345be9c78b1f4c93(SB)
	0x0032 00050 (sample/sample.go:46)	FUNCDATA	$2, gclocals·9fb7f0986f647f17cb53dda1484e0f7a(SB)
	0x0032 00050 (sample/sample.go:47)	PCDATA	$0, $-1
	0x0032 00050 (sample/sample.go:47)	PCDATA	$1, $-1
	0x0032 00050 (sample/sample.go:47)	JMP	52
	0x0034 00052 (sample/sample.go:48)	PCDATA	$0, $0
	0x0034 00052 (sample/sample.go:48)	PCDATA	$1, $1
	0x0034 00052 (sample/sample.go:48)	MOVQ	$0, "".a+424(SP)
	0x0040 00064 (sample/sample.go:48)	XORPS	X0, X0
	0x0043 00067 (sample/sample.go:48)	MOVUPS	X0, "".a+432(SP)
	0x004b 00075 (sample/sample.go:48)	MOVQ	$0, "".~r1+304(SP)
	0x0057 00087 (sample/sample.go:48)	XORPS	X0, X0
	0x005a 00090 (sample/sample.go:48)	MOVUPS	X0, "".~r1+312(SP)
	0x0062 00098 (<unknown line number>)	NOP
	0x0062 00098 (sample/sample.go:15)	PCDATA	$0, $1
	0x0062 00098 (sample/sample.go:15)	MOVQ	"".a+424(SP), AX
	0x006a 00106 (sample/sample.go:15)	MOVQ	"".a+440(SP), CX
	0x0072 00114 (sample/sample.go:15)	PCDATA	$1, $0
	0x0072 00114 (sample/sample.go:15)	MOVQ	"".a+432(SP), DX
	0x007a 00122 (sample/sample.go:15)	MOVQ	AX, "".b+376(SP)
	0x0082 00130 (sample/sample.go:15)	MOVQ	DX, "".b+384(SP)
	0x008a 00138 (sample/sample.go:15)	MOVQ	CX, "".b+392(SP)
	0x0092 00146 (sample/sample.go:48)	PCDATA	$0, $0
	0x0092 00146 (sample/sample.go:48)	MOVQ	AX, "".~r1+304(SP)
	0x009a 00154 (sample/sample.go:48)	MOVQ	DX, "".~r1+312(SP)
	0x00a2 00162 (sample/sample.go:48)	MOVQ	CX, "".~r1+320(SP)
	0x00aa 00170 (sample/sample.go:48)	JMP	172
	0x00ac 00172 (sample/sample.go:49)	MOVW	$0, ""..autotmp_26+61(SP)
	0x00b3 00179 (sample/sample.go:49)	MOVB	$0, ""..autotmp_26+63(SP)
	0x00b8 00184 (sample/sample.go:49)	PCDATA	$0, $1
	0x00b8 00184 (sample/sample.go:49)	LEAQ	""..autotmp_26+61(SP), AX
	0x00bd 00189 (sample/sample.go:49)	PCDATA	$1, $2
	0x00bd 00189 (sample/sample.go:49)	MOVQ	AX, ""..autotmp_24+232(SP)
	0x00c5 00197 (sample/sample.go:49)	PCDATA	$0, $0
	0x00c5 00197 (sample/sample.go:49)	TESTB	AL, (AX)
	0x00c7 00199 (sample/sample.go:49)	MOVB	$97, ""..autotmp_26+61(SP)
	0x00cc 00204 (sample/sample.go:49)	PCDATA	$0, $1
	0x00cc 00204 (sample/sample.go:49)	MOVQ	""..autotmp_24+232(SP), AX
	0x00d4 00212 (sample/sample.go:49)	TESTB	AL, (AX)
	0x00d6 00214 (sample/sample.go:49)	PCDATA	$0, $0
	0x00d6 00214 (sample/sample.go:49)	MOVB	$98, 1(AX)
	0x00da 00218 (sample/sample.go:49)	PCDATA	$0, $1
	0x00da 00218 (sample/sample.go:49)	MOVQ	""..autotmp_24+232(SP), AX
	0x00e2 00226 (sample/sample.go:49)	TESTB	AL, (AX)
	0x00e4 00228 (sample/sample.go:49)	PCDATA	$0, $0
	0x00e4 00228 (sample/sample.go:49)	MOVB	$99, 2(AX)
	0x00e8 00232 (sample/sample.go:49)	PCDATA	$0, $1
	0x00e8 00232 (sample/sample.go:49)	PCDATA	$1, $0
	0x00e8 00232 (sample/sample.go:49)	MOVQ	""..autotmp_24+232(SP), AX
	0x00f0 00240 (sample/sample.go:49)	TESTB	AL, (AX)
	0x00f2 00242 (sample/sample.go:49)	JMP	244
	0x00f4 00244 (sample/sample.go:49)	MOVQ	AX, ""..autotmp_23+448(SP)
	0x00fc 00252 (sample/sample.go:49)	MOVQ	$3, ""..autotmp_23+456(SP)
	0x0108 00264 (sample/sample.go:49)	MOVQ	$3, ""..autotmp_23+464(SP)
	0x0114 00276 (sample/sample.go:49)	PCDATA	$0, $0
	0x0114 00276 (sample/sample.go:49)	PCDATA	$1, $3
	0x0114 00276 (sample/sample.go:49)	MOVQ	AX, "".a+400(SP)
	0x011c 00284 (sample/sample.go:49)	MOVQ	$3, "".a+408(SP)
	0x0128 00296 (sample/sample.go:49)	MOVQ	$3, "".a+416(SP)
	0x0134 00308 (sample/sample.go:49)	MOVQ	$0, "".~r1+328(SP)
	0x0140 00320 (sample/sample.go:49)	XORPS	X0, X0
	0x0143 00323 (sample/sample.go:49)	MOVUPS	X0, "".~r1+336(SP)
	0x014b 00331 (<unknown line number>)	NOP
	0x014b 00331 (sample/sample.go:15)	PCDATA	$0, $1
	0x014b 00331 (sample/sample.go:15)	MOVQ	"".a+400(SP), AX
	0x0153 00339 (sample/sample.go:15)	MOVQ	"".a+408(SP), CX
	0x015b 00347 (sample/sample.go:15)	PCDATA	$1, $0
	0x015b 00347 (sample/sample.go:15)	MOVQ	"".a+416(SP), DX
	0x0163 00355 (sample/sample.go:15)	MOVQ	AX, "".b+352(SP)
	0x016b 00363 (sample/sample.go:15)	MOVQ	CX, "".b+360(SP)
	0x0173 00371 (sample/sample.go:15)	MOVQ	DX, "".b+368(SP)
	0x017b 00379 (sample/sample.go:49)	PCDATA	$0, $0
	0x017b 00379 (sample/sample.go:49)	MOVQ	AX, "".~r1+328(SP)
	0x0183 00387 (sample/sample.go:49)	MOVQ	CX, "".~r1+336(SP)
	0x018b 00395 (sample/sample.go:49)	MOVQ	DX, "".~r1+344(SP)
	0x0193 00403 (sample/sample.go:49)	JMP	405
	0x0195 00405 (sample/sample.go:50)	MOVQ	$1, "".x+152(SP)
	0x01a1 00417 (sample/sample.go:50)	JMP	419
	0x01a3 00419 (sample/sample.go:51)	MOVQ	$2, "".x+136(SP)
	0x01af 00431 (sample/sample.go:51)	MOVQ	$0, "".~r1+88(SP)
	0x01b8 00440 (sample/sample.go:51)	MOVQ	"".x+136(SP), AX
	0x01c0 00448 (sample/sample.go:51)	MOVQ	AX, "".~r1+88(SP)
	0x01c5 00453 (sample/sample.go:51)	JMP	455
	0x01c7 00455 (sample/sample.go:52)	MOVQ	$5, "".x+128(SP)
	0x01d3 00467 (sample/sample.go:52)	MOVQ	$7, "".y+112(SP)
	0x01dc 00476 (sample/sample.go:52)	MOVQ	$0, "".~r2+72(SP)
	0x01e5 00485 (sample/sample.go:52)	MOVQ	$0, "".~r3+64(SP)
	0x01ee 00494 (<unknown line number>)	NOP
	0x01ee 00494 (sample/sample.go:28)	MOVQ	"".x+128(SP), AX
	0x01f6 00502 (sample/sample.go:28)	ADDQ	"".y+112(SP), AX
	0x01fb 00507 (sample/sample.go:52)	MOVQ	AX, ""..autotmp_27+192(SP)
	0x0203 00515 (sample/sample.go:28)	MOVQ	"".x+128(SP), AX
	0x020b 00523 (sample/sample.go:28)	MOVQ	"".y+112(SP), CX
	0x0210 00528 (sample/sample.go:28)	IMULQ	CX, AX
	0x0214 00532 (sample/sample.go:52)	MOVQ	AX, ""..autotmp_28+184(SP)
	0x021c 00540 (sample/sample.go:52)	MOVQ	""..autotmp_27+192(SP), AX
	0x0224 00548 (sample/sample.go:52)	MOVQ	AX, "".~r2+72(SP)
	0x0229 00553 (sample/sample.go:52)	MOVQ	""..autotmp_28+184(SP), AX
	0x0231 00561 (sample/sample.go:52)	MOVQ	AX, "".~r3+64(SP)
	0x0236 00566 (sample/sample.go:52)	JMP	568
	0x0238 00568 (sample/sample.go:53)	MOVQ	$5, "".x+120(SP)
	0x0241 00577 (sample/sample.go:53)	MOVQ	$7, "".y+104(SP)
	0x024a 00586 (sample/sample.go:53)	MOVQ	$0, "".sum+160(SP)
	0x0256 00598 (sample/sample.go:53)	MOVQ	$0, "".mul+168(SP)
	0x0262 00610 (<unknown line number>)	NOP
	0x0262 00610 (sample/sample.go:32)	MOVQ	"".x+120(SP), AX
	0x0267 00615 (sample/sample.go:32)	ADDQ	"".y+104(SP), AX
	0x026c 00620 (sample/sample.go:32)	MOVQ	AX, "".sum+160(SP)
	0x0274 00628 (sample/sample.go:33)	MOVQ	"".y+104(SP), AX
	0x0279 00633 (sample/sample.go:33)	MOVQ	"".x+120(SP), CX
	0x027e 00638 (sample/sample.go:33)	IMULQ	AX, CX
	0x0282 00642 (sample/sample.go:33)	MOVQ	CX, "".mul+168(SP)
	0x028a 00650 (sample/sample.go:53)	JMP	652
	0x028c 00652 (sample/sample.go:54)	MOVQ	$2, "".x+144(SP)
	0x0298 00664 (sample/sample.go:54)	MOVQ	$3, "".y+96(SP)
	0x02a1 00673 (sample/sample.go:54)	MOVQ	$0, "".~r2+80(SP)
	0x02aa 00682 (<unknown line number>)	NOP
	0x02aa 00682 (sample/sample.go:39)	MOVQ	"".x+144(SP), AX
	0x02b2 00690 (sample/sample.go:39)	ADDQ	"".y+96(SP), AX
	0x02b7 00695 (sample/sample.go:54)	MOVQ	AX, ""..autotmp_29+176(SP)
	0x02bf 00703 (sample/sample.go:54)	MOVQ	AX, "".~r2+80(SP)
	0x02c4 00708 (sample/sample.go:54)	JMP	710
	0x02c6 00710 (sample/sample.go:55)	PCDATA	$0, $1
	0x02c6 00710 (sample/sample.go:55)	PCDATA	$1, $4
	0x02c6 00710 (sample/sample.go:55)	LEAQ	go.string."hello"(SB), AX
	0x02cd 00717 (sample/sample.go:55)	PCDATA	$0, $0
	0x02cd 00717 (sample/sample.go:55)	MOVQ	AX, "".x+272(SP)
	0x02d5 00725 (sample/sample.go:55)	MOVQ	$5, "".x+280(SP)
	0x02e1 00737 (sample/sample.go:55)	PCDATA	$0, $1
	0x02e1 00737 (sample/sample.go:55)	PCDATA	$1, $5
	0x02e1 00737 (sample/sample.go:55)	LEAQ	go.string." world"(SB), AX
	0x02e8 00744 (sample/sample.go:55)	PCDATA	$0, $0
	0x02e8 00744 (sample/sample.go:55)	MOVQ	AX, "".y+256(SP)
	0x02f0 00752 (sample/sample.go:55)	MOVQ	$6, "".y+264(SP)
	0x02fc 00764 (sample/sample.go:55)	XORPS	X0, X0
	0x02ff 00767 (sample/sample.go:55)	MOVUPS	X0, "".~r2+240(SP)
	0x0307 00775 (<unknown line number>)	NOP
	0x0307 00775 (sample/sample.go:43)	PCDATA	$0, $1
	0x0307 00775 (sample/sample.go:43)	LEAQ	""..autotmp_31+200(SP), AX
	0x030f 00783 (sample/sample.go:43)	PCDATA	$0, $0
	0x030f 00783 (sample/sample.go:43)	MOVQ	AX, (SP)
	0x0313 00787 (sample/sample.go:43)	PCDATA	$0, $1
	0x0313 00787 (sample/sample.go:43)	MOVQ	"".x+272(SP), AX
	0x031b 00795 (sample/sample.go:43)	PCDATA	$1, $6
	0x031b 00795 (sample/sample.go:43)	MOVQ	"".x+280(SP), CX
	0x0323 00803 (sample/sample.go:43)	PCDATA	$0, $0
	0x0323 00803 (sample/sample.go:43)	MOVQ	AX, 8(SP)
	0x0328 00808 (sample/sample.go:43)	MOVQ	CX, 16(SP)
	0x032d 00813 (sample/sample.go:43)	PCDATA	$0, $1
	0x032d 00813 (sample/sample.go:43)	MOVQ	"".y+256(SP), AX
	0x0335 00821 (sample/sample.go:43)	PCDATA	$1, $0
	0x0335 00821 (sample/sample.go:43)	MOVQ	"".y+264(SP), CX
	0x033d 00829 (sample/sample.go:43)	PCDATA	$0, $0
	0x033d 00829 (sample/sample.go:43)	MOVQ	AX, 24(SP)
	0x0342 00834 (sample/sample.go:43)	MOVQ	CX, 32(SP)
	0x0347 00839 (sample/sample.go:43)	CALL	runtime.concatstring2(SB)
	0x034c 00844 (sample/sample.go:43)	PCDATA	$0, $1
	0x034c 00844 (sample/sample.go:43)	MOVQ	40(SP), AX
	0x0351 00849 (sample/sample.go:43)	MOVQ	48(SP), CX
	0x0356 00854 (sample/sample.go:55)	MOVQ	AX, ""..autotmp_30+288(SP)
	0x035e 00862 (sample/sample.go:55)	MOVQ	CX, ""..autotmp_30+296(SP)
	0x0366 00870 (sample/sample.go:55)	PCDATA	$0, $0
	0x0366 00870 (sample/sample.go:55)	MOVQ	AX, "".~r2+240(SP)
	0x036e 00878 (sample/sample.go:55)	MOVQ	CX, "".~r2+248(SP)
	0x0376 00886 (sample/sample.go:55)	JMP	888
	0x0378 00888 (sample/sample.go:55)	PCDATA	$0, $-1
	0x0378 00888 (sample/sample.go:55)	PCDATA	$1, $-1
	0x0378 00888 (sample/sample.go:55)	MOVQ	472(SP), BP
	0x0380 00896 (sample/sample.go:55)	ADDQ	$480, SP
	0x0387 00903 (sample/sample.go:55)	RET
	0x0388 00904 (sample/sample.go:55)	NOP
	0x0388 00904 (sample/sample.go:46)	PCDATA	$1, $-1
	0x0388 00904 (sample/sample.go:46)	PCDATA	$0, $-2
	0x0388 00904 (sample/sample.go:46)	CALL	runtime.morestack_noctxt(SB)
	0x038d 00909 (sample/sample.go:46)	PCDATA	$0, $-1
	0x038d 00909 (sample/sample.go:46)	JMP	0
	0x0000 64 48 8b 0c 25 00 00 00 00 48 8d 84 24 a0 fe ff  dH..%....H..$...
	0x0010 ff 48 3b 41 10 0f 86 6d 03 00 00 48 81 ec e0 01  .H;A...m...H....
	0x0020 00 00 48 89 ac 24 d8 01 00 00 48 8d ac 24 d8 01  ..H..$....H..$..
	0x0030 00 00 eb 00 48 c7 84 24 a8 01 00 00 00 00 00 00  ....H..$........
	0x0040 0f 57 c0 0f 11 84 24 b0 01 00 00 48 c7 84 24 30  .W....$....H..$0
	0x0050 01 00 00 00 00 00 00 0f 57 c0 0f 11 84 24 38 01  ........W....$8.
	0x0060 00 00 48 8b 84 24 a8 01 00 00 48 8b 8c 24 b8 01  ..H..$....H..$..
	0x0070 00 00 48 8b 94 24 b0 01 00 00 48 89 84 24 78 01  ..H..$....H..$x.
	0x0080 00 00 48 89 94 24 80 01 00 00 48 89 8c 24 88 01  ..H..$....H..$..
	0x0090 00 00 48 89 84 24 30 01 00 00 48 89 94 24 38 01  ..H..$0...H..$8.
	0x00a0 00 00 48 89 8c 24 40 01 00 00 eb 00 66 c7 44 24  ..H..$@.....f.D$
	0x00b0 3d 00 00 c6 44 24 3f 00 48 8d 44 24 3d 48 89 84  =...D$?.H.D$=H..
	0x00c0 24 e8 00 00 00 84 00 c6 44 24 3d 61 48 8b 84 24  $.......D$=aH..$
	0x00d0 e8 00 00 00 84 00 c6 40 01 62 48 8b 84 24 e8 00  .......@.bH..$..
	0x00e0 00 00 84 00 c6 40 02 63 48 8b 84 24 e8 00 00 00  .....@.cH..$....
	0x00f0 84 00 eb 00 48 89 84 24 c0 01 00 00 48 c7 84 24  ....H..$....H..$
	0x0100 c8 01 00 00 03 00 00 00 48 c7 84 24 d0 01 00 00  ........H..$....
	0x0110 03 00 00 00 48 89 84 24 90 01 00 00 48 c7 84 24  ....H..$....H..$
	0x0120 98 01 00 00 03 00 00 00 48 c7 84 24 a0 01 00 00  ........H..$....
	0x0130 03 00 00 00 48 c7 84 24 48 01 00 00 00 00 00 00  ....H..$H.......
	0x0140 0f 57 c0 0f 11 84 24 50 01 00 00 48 8b 84 24 90  .W....$P...H..$.
	0x0150 01 00 00 48 8b 8c 24 98 01 00 00 48 8b 94 24 a0  ...H..$....H..$.
	0x0160 01 00 00 48 89 84 24 60 01 00 00 48 89 8c 24 68  ...H..$`...H..$h
	0x0170 01 00 00 48 89 94 24 70 01 00 00 48 89 84 24 48  ...H..$p...H..$H
	0x0180 01 00 00 48 89 8c 24 50 01 00 00 48 89 94 24 58  ...H..$P...H..$X
	0x0190 01 00 00 eb 00 48 c7 84 24 98 00 00 00 01 00 00  .....H..$.......
	0x01a0 00 eb 00 48 c7 84 24 88 00 00 00 02 00 00 00 48  ...H..$........H
	0x01b0 c7 44 24 58 00 00 00 00 48 8b 84 24 88 00 00 00  .D$X....H..$....
	0x01c0 48 89 44 24 58 eb 00 48 c7 84 24 80 00 00 00 05  H.D$X..H..$.....
	0x01d0 00 00 00 48 c7 44 24 70 07 00 00 00 48 c7 44 24  ...H.D$p....H.D$
	0x01e0 48 00 00 00 00 48 c7 44 24 40 00 00 00 00 48 8b  H....H.D$@....H.
	0x01f0 84 24 80 00 00 00 48 03 44 24 70 48 89 84 24 c0  .$....H.D$pH..$.
	0x0200 00 00 00 48 8b 84 24 80 00 00 00 48 8b 4c 24 70  ...H..$....H.L$p
	0x0210 48 0f af c1 48 89 84 24 b8 00 00 00 48 8b 84 24  H...H..$....H..$
	0x0220 c0 00 00 00 48 89 44 24 48 48 8b 84 24 b8 00 00  ....H.D$HH..$...
	0x0230 00 48 89 44 24 40 eb 00 48 c7 44 24 78 05 00 00  .H.D$@..H.D$x...
	0x0240 00 48 c7 44 24 68 07 00 00 00 48 c7 84 24 a0 00  .H.D$h....H..$..
	0x0250 00 00 00 00 00 00 48 c7 84 24 a8 00 00 00 00 00  ......H..$......
	0x0260 00 00 48 8b 44 24 78 48 03 44 24 68 48 89 84 24  ..H.D$xH.D$hH..$
	0x0270 a0 00 00 00 48 8b 44 24 68 48 8b 4c 24 78 48 0f  ....H.D$hH.L$xH.
	0x0280 af c8 48 89 8c 24 a8 00 00 00 eb 00 48 c7 84 24  ..H..$......H..$
	0x0290 90 00 00 00 02 00 00 00 48 c7 44 24 60 03 00 00  ........H.D$`...
	0x02a0 00 48 c7 44 24 50 00 00 00 00 48 8b 84 24 90 00  .H.D$P....H..$..
	0x02b0 00 00 48 03 44 24 60 48 89 84 24 b0 00 00 00 48  ..H.D$`H..$....H
	0x02c0 89 44 24 50 eb 00 48 8d 05 00 00 00 00 48 89 84  .D$P..H......H..
	0x02d0 24 10 01 00 00 48 c7 84 24 18 01 00 00 05 00 00  $....H..$.......
	0x02e0 00 48 8d 05 00 00 00 00 48 89 84 24 00 01 00 00  .H......H..$....
	0x02f0 48 c7 84 24 08 01 00 00 06 00 00 00 0f 57 c0 0f  H..$.........W..
	0x0300 11 84 24 f0 00 00 00 48 8d 84 24 c8 00 00 00 48  ..$....H..$....H
	0x0310 89 04 24 48 8b 84 24 10 01 00 00 48 8b 8c 24 18  ..$H..$....H..$.
	0x0320 01 00 00 48 89 44 24 08 48 89 4c 24 10 48 8b 84  ...H.D$.H.L$.H..
	0x0330 24 00 01 00 00 48 8b 8c 24 08 01 00 00 48 89 44  $....H..$....H.D
	0x0340 24 18 48 89 4c 24 20 e8 00 00 00 00 48 8b 44 24  $.H.L$ .....H.D$
	0x0350 28 48 8b 4c 24 30 48 89 84 24 20 01 00 00 48 89  (H.L$0H..$ ...H.
	0x0360 8c 24 28 01 00 00 48 89 84 24 f0 00 00 00 48 89  .$(...H..$....H.
	0x0370 8c 24 f8 00 00 00 eb 00 48 8b ac 24 d8 01 00 00  .$......H..$....
	0x0380 48 81 c4 e0 01 00 00 c3 e8 00 00 00 00 e9 6e fc  H.............n.
	0x0390 ff ff                                            ..
	rel 5+4 t=17 TLS+0
	rel 713+4 t=16 go.string."hello"+0
	rel 740+4 t=16 go.string." world"+0
	rel 840+4 t=8 runtime.concatstring2+0
	rel 905+4 t=8 runtime.morestack_noctxt+0
go.cuinfo.packagename. SDWARFINFO dupok size=0
	0x0000 6d 61 69 6e                                      main
go.info."".min$abstract SDWARFINFO dupok size=9
	0x0000 04 2e 6d 69 6e 00 01 01 00                       ..min....
go.info."".slice$abstract SDWARFINFO dupok size=27
	0x0000 04 2e 73 6c 69 63 65 00 01 01 11 61 00 00 00 00  ..slice....a....
	0x0010 00 00 0c 62 00 0f 00 00 00 00 00                 ...b.......
	rel 14+4 t=29 go.info.[]uint8+0
	rel 22+4 t=29 go.info.[]uint8+0
go.info."".arg1$abstract SDWARFINFO dupok size=18
	0x0000 04 2e 61 72 67 31 00 01 01 11 78 00 00 00 00 00  ..arg1....x.....
	0x0010 00 00                                            ..
	rel 13+4 t=29 go.info.int+0
go.info."".arg1ret1$abstract SDWARFINFO dupok size=22
	0x0000 04 2e 61 72 67 31 72 65 74 31 00 01 01 11 78 00  ..arg1ret1....x.
	0x0010 00 00 00 00 00 00                                ......
	rel 17+4 t=29 go.info.int+0
go.info."".sumAndMul$abstract SDWARFINFO dupok size=31
	0x0000 04 2e 73 75 6d 41 6e 64 4d 75 6c 00 01 01 11 78  ..sumAndMul....x
	0x0010 00 00 00 00 00 00 11 79 00 00 00 00 00 00 00     .......y.......
	rel 18+4 t=29 go.info.int+0
	rel 26+4 t=29 go.info.int+0
go.info."".sumAndMulWithNamedReturn$abstract SDWARFINFO dupok size=66
	0x0000 04 2e 73 75 6d 41 6e 64 4d 75 6c 57 69 74 68 4e  ..sumAndMulWithN
	0x0010 61 6d 65 64 52 65 74 75 72 6e 00 01 01 11 78 00  amedReturn....x.
	0x0020 00 00 00 00 00 11 79 00 00 00 00 00 00 11 73 75  ......y.......su
	0x0030 6d 00 01 00 00 00 00 11 6d 75 6c 00 01 00 00 00  m.......mul.....
	0x0040 00 00                                            ..
	rel 33+4 t=29 go.info.int+0
	rel 41+4 t=29 go.info.int+0
	rel 51+4 t=29 go.info.int+0
	rel 61+4 t=29 go.info.int+0
go.info."".sum$abstract SDWARFINFO dupok size=25
	0x0000 04 2e 73 75 6d 00 01 01 11 78 00 00 00 00 00 00  ..sum....x......
	0x0010 11 79 00 00 00 00 00 00 00                       .y.......
	rel 12+4 t=29 go.info.int+0
	rel 20+4 t=29 go.info.int+0
go.info."".concate$abstract SDWARFINFO dupok size=29
	0x0000 04 2e 63 6f 6e 63 61 74 65 00 01 01 11 78 00 00  ..concate....x..
	0x0010 00 00 00 00 11 79 00 00 00 00 00 00 00           .....y.......
	rel 16+4 t=29 go.info.string+0
	rel 24+4 t=29 go.info.string+0
go.loc."".min SDWARFLOC size=0
go.info."".min SDWARFINFO size=24
	0x0000 05 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0010 00 00 00 00 00 01 9c 00                          ........
	rel 1+4 t=29 go.info."".min$abstract+0
	rel 5+8 t=1 "".min+0
	rel 13+8 t=1 "".min+1
go.range."".min SDWARFRANGE size=0
go.debuglines."".min SDWARFMISC size=8
	0x0000 04 02 14 04 01 03 7b 01                          ......{.
go.loc."".char SDWARFLOC size=0
go.info."".char SDWARFINFO size=69
	0x0000 03 22 22 2e 63 68 61 72 00 00 00 00 00 00 00 00  ."".char........
	0x0010 00 00 00 00 00 00 00 00 00 01 9c 00 00 00 00 01  ................
	0x0020 0a 62 00 09 00 00 00 00 02 91 6f 0f 61 00 00 08  .b........o.a...
	0x0030 00 00 00 00 01 9c 0f 7e 72 31 00 01 08 00 00 00  .......~r1......
	0x0040 00 02 91 08 00                                   .....
	rel 0+0 t=24 type.uint8+0
	rel 9+8 t=1 "".char+0
	rel 17+8 t=1 "".char+48
	rel 27+4 t=30 gofile../root/go/src/github.com/DQNEO/babygo/sample/sample.go+0
	rel 36+4 t=29 go.info.uint8+0
	rel 48+4 t=29 go.info.uint8+0
	rel 61+4 t=29 go.info.uint8+0
go.range."".char SDWARFRANGE size=0
go.debuglines."".char SDWARFMISC size=20
	0x0000 04 02 0a 03 02 14 ce 06 41 06 38 42 06 41 04 01  ........A.8B.A..
	0x0010 03 76 06 01                                      .v..
go.loc."".slice SDWARFLOC size=0
go.info."".slice SDWARFINFO size=53
	0x0000 05 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0010 00 00 00 00 00 01 9c 12 00 00 00 00 01 9c 0f 7e  ...............~
	0x0020 72 31 00 01 0e 00 00 00 00 02 91 18 0d 00 00 00  r1..............
	0x0030 00 02 91 58 00                                   ...X.
	rel 0+0 t=24 type.[]uint8+0
	rel 1+4 t=29 go.info."".slice$abstract+0
	rel 5+8 t=1 "".slice+0
	rel 13+8 t=1 "".slice+85
	rel 24+4 t=29 go.info."".slice$abstract+10
	rel 37+4 t=29 go.info.[]uint8+0
	rel 45+4 t=29 go.info."".slice$abstract+18
go.range."".slice SDWARFRANGE size=0
go.debuglines."".slice SDWARFMISC size=23
	0x0000 04 02 0a 03 08 14 06 f5 06 60 06 41 06 08 10 06  .........`.A....
	0x0010 41 04 01 03 71 06 01                             A...q..
go.loc."".arg1 SDWARFLOC size=0
go.info."".arg1 SDWARFINFO size=31
	0x0000 05 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0010 00 00 00 00 00 01 9c 12 00 00 00 00 01 9c 00     ...............
	rel 0+0 t=24 type.int+0
	rel 1+4 t=29 go.info."".arg1$abstract+0
	rel 5+8 t=1 "".arg1+0
	rel 13+8 t=1 "".arg1+1
	rel 24+4 t=29 go.info."".arg1$abstract+9
go.range."".arg1 SDWARFRANGE size=0
go.debuglines."".arg1 SDWARFMISC size=10
	0x0000 04 02 03 0f 14 04 01 03 6c 01                    ........l.
go.loc."".arg1ret1 SDWARFLOC size=0
go.info."".arg1ret1 SDWARFINFO size=45
	0x0000 05 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0010 00 00 00 00 00 01 9c 12 00 00 00 00 01 9c 0f 7e  ...............~
	0x0020 72 31 00 01 17 00 00 00 00 02 91 08 00           r1...........
	rel 0+0 t=24 type.int+0
	rel 1+4 t=29 go.info."".arg1ret1$abstract+0
	rel 5+8 t=1 "".arg1ret1+0
	rel 13+8 t=1 "".arg1ret1+20
	rel 24+4 t=29 go.info."".arg1ret1$abstract+13
	rel 37+4 t=29 go.info.int+0
go.range."".arg1ret1 SDWARFRANGE size=0
go.debuglines."".arg1ret1 SDWARFMISC size=14
	0x0000 04 02 03 11 14 6a 06 41 04 01 03 69 06 01        .....j.A...i..
go.loc."".sumAndMul SDWARFLOC size=0
go.info."".sumAndMul SDWARFINFO size=67
	0x0000 05 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0010 00 00 00 00 00 01 9c 12 00 00 00 00 01 9c 12 00  ................
	0x0020 00 00 00 02 91 08 0f 7e 72 32 00 01 1b 00 00 00  .......~r2......
	0x0030 00 02 91 10 0f 7e 72 33 00 01 1b 00 00 00 00 02  .....~r3........
	0x0040 91 18 00                                         ...
	rel 0+0 t=24 type.int+0
	rel 1+4 t=29 go.info."".sumAndMul$abstract+0
	rel 5+8 t=1 "".sumAndMul+0
	rel 13+8 t=1 "".sumAndMul+53
	rel 24+4 t=29 go.info."".sumAndMul$abstract+14
	rel 31+4 t=29 go.info."".sumAndMul$abstract+22
	rel 45+4 t=29 go.info.int+0
	rel 59+4 t=29 go.info.int+0
go.range."".sumAndMul SDWARFRANGE size=0
go.debuglines."".sumAndMul SDWARFMISC size=17
	0x0000 04 02 03 15 14 06 69 06 6a 06 41 04 01 03 65 06  ......i.j.A...e.
	0x0010 01                                               .
go.loc."".sumAndMulWithNamedReturn SDWARFLOC size=0
go.info."".sumAndMulWithNamedReturn SDWARFINFO size=55
	0x0000 05 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0010 00 00 00 00 00 01 9c 12 00 00 00 00 01 9c 12 00  ................
	0x0020 00 00 00 02 91 08 12 00 00 00 00 02 91 10 12 00  ................
	0x0030 00 00 00 02 91 18 00                             .......
	rel 0+0 t=24 type.int+0
	rel 1+4 t=29 go.info."".sumAndMulWithNamedReturn$abstract+0
	rel 5+8 t=1 "".sumAndMulWithNamedReturn+0
	rel 13+8 t=1 "".sumAndMulWithNamedReturn+53
	rel 24+4 t=29 go.info."".sumAndMulWithNamedReturn$abstract+29
	rel 31+4 t=29 go.info."".sumAndMulWithNamedReturn$abstract+37
	rel 39+4 t=29 go.info."".sumAndMulWithNamedReturn$abstract+45
	rel 47+4 t=29 go.info."".sumAndMulWithNamedReturn$abstract+55
go.range."".sumAndMulWithNamedReturn SDWARFRANGE size=0
go.debuglines."".sumAndMulWithNamedReturn SDWARFMISC size=22
	0x0000 04 02 03 19 14 06 69 06 6a 06 41 06 74 06 41 06  ......i.j.A.t.A.
	0x0010 9c 04 01 03 5f 01                                ...._.
go.loc."".sum SDWARFLOC size=0
go.info."".sum SDWARFINFO size=53
	0x0000 05 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0010 00 00 00 00 00 01 9c 12 00 00 00 00 01 9c 12 00  ................
	0x0020 00 00 00 02 91 08 0f 7e 72 32 00 01 26 00 00 00  .......~r2..&...
	0x0030 00 02 91 10 00                                   .....
	rel 0+0 t=24 type.int+0
	rel 1+4 t=29 go.info."".sum$abstract+0
	rel 5+8 t=1 "".sum+0
	rel 13+8 t=1 "".sum+25
	rel 24+4 t=29 go.info."".sum$abstract+8
	rel 31+4 t=29 go.info."".sum$abstract+16
	rel 45+4 t=29 go.info.int+0
go.range."".sum SDWARFRANGE size=0
go.debuglines."".sum SDWARFMISC size=14
	0x0000 04 02 03 20 14 6a 06 41 04 01 03 5a 06 01        ... .j.A...Z..
go.loc."".concate SDWARFLOC size=0
go.info."".concate SDWARFINFO size=53
	0x0000 05 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0010 00 00 00 00 00 01 9c 12 00 00 00 00 01 9c 12 00  ................
	0x0020 00 00 00 02 91 10 0f 7e 72 32 00 01 2a 00 00 00  .......~r2..*...
	0x0030 00 02 91 20 00                                   ... .
	rel 0+0 t=24 type.string+0
	rel 1+4 t=29 go.info."".concate$abstract+0
	rel 5+8 t=1 "".concate+0
	rel 13+8 t=1 "".concate+127
	rel 24+4 t=29 go.info."".concate$abstract+12
	rel 31+4 t=29 go.info."".concate$abstract+20
	rel 45+4 t=29 go.info.string+0
go.range."".concate SDWARFRANGE size=0
go.debuglines."".concate SDWARFMISC size=26
	0x0000 04 02 03 24 14 0a a5 06 b9 06 42 06 5f 06 08 af  ...$......B._...
	0x0010 06 41 06 08 4a 04 01 03 57 01                    .A..J...W.
go.string."hello" SRODATA dupok size=5
	0x0000 68 65 6c 6c 6f                                   hello
go.string." world" SRODATA dupok size=6
	0x0000 20 77 6f 72 6c 64                                 world
go.loc."".main SDWARFLOC size=0
go.info."".main SDWARFINFO size=308
	0x0000 03 22 22 2e 6d 61 69 6e 00 00 00 00 00 00 00 00  ."".main........
	0x0010 00 00 00 00 00 00 00 00 00 01 9c 00 00 00 00 01  ................
	0x0020 06 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0030 00 00 00 00 00 00 00 00 00 30 12 00 00 00 00 02  .........0......
	0x0040 91 40 0d 00 00 00 00 03 91 90 7f 00 06 00 00 00  .@..............
	0x0050 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0060 00 00 00 00 00 31 12 00 00 00 00 03 91 a8 7f 0d  .....1..........
	0x0070 00 00 00 00 03 91 f8 7e 00 07 00 00 00 00 00 00  .......~........
	0x0080 00 00 00 00 00 00 34 12 00 00 00 00 03 91 98 7d  ......4........}
	0x0090 12 00 00 00 00 03 91 88 7d 00 06 00 00 00 00 00  ........}.......
	0x00a0 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x00b0 00 00 00 35 12 00 00 00 00 03 91 90 7d 12 00 00  ...5........}...
	0x00c0 00 00 03 91 80 7d 12 00 00 00 00 03 91 b8 7d 12  .....}........}.
	0x00d0 00 00 00 00 03 91 c0 7d 00 06 00 00 00 00 00 00  .......}........
	0x00e0 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x00f0 00 00 36 12 00 00 00 00 03 91 a8 7d 12 00 00 00  ..6........}....
	0x0100 00 03 91 f8 7c 00 06 00 00 00 00 00 00 00 00 00  ....|...........
	0x0110 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 37  ...............7
	0x0120 12 00 00 00 00 03 91 a8 7e 12 00 00 00 00 03 91  ........~.......
	0x0130 98 7e 00 00                                      .~..
	rel 0+0 t=24 type.*[3]uint8+0
	rel 0+0 t=24 type.[32]uint8+0
	rel 0+0 t=24 type.[3]uint8+0
	rel 0+0 t=24 type.[]uint8+0
	rel 0+0 t=24 type.int+0
	rel 0+0 t=24 type.string+0
	rel 9+8 t=1 "".main+0
	rel 17+8 t=1 "".main+914
	rel 27+4 t=30 gofile../root/go/src/github.com/DQNEO/babygo/sample/sample.go+0
	rel 33+4 t=29 go.info."".slice$abstract+0
	rel 37+8 t=1 "".main+98
	rel 45+8 t=1 "".main+146
	rel 53+4 t=30 gofile../root/go/src/github.com/DQNEO/babygo/sample/sample.go+0
	rel 59+4 t=29 go.info."".slice$abstract+10
	rel 67+4 t=29 go.info."".slice$abstract+18
	rel 77+4 t=29 go.info."".slice$abstract+0
	rel 81+8 t=1 "".main+331
	rel 89+8 t=1 "".main+379
	rel 97+4 t=30 gofile../root/go/src/github.com/DQNEO/babygo/sample/sample.go+0
	rel 103+4 t=29 go.info."".slice$abstract+10
	rel 112+4 t=29 go.info."".slice$abstract+18
	rel 122+4 t=29 go.info."".sumAndMul$abstract+0
	rel 126+4 t=29 go.range."".main+0
	rel 130+4 t=30 gofile../root/go/src/github.com/DQNEO/babygo/sample/sample.go+0
	rel 136+4 t=29 go.info."".sumAndMul$abstract+14
	rel 145+4 t=29 go.info."".sumAndMul$abstract+22
	rel 155+4 t=29 go.info."".sumAndMulWithNamedReturn$abstract+0
	rel 159+8 t=1 "".main+610
	rel 167+8 t=1 "".main+650
	rel 175+4 t=30 gofile../root/go/src/github.com/DQNEO/babygo/sample/sample.go+0
	rel 181+4 t=29 go.info."".sumAndMulWithNamedReturn$abstract+29
	rel 190+4 t=29 go.info."".sumAndMulWithNamedReturn$abstract+37
	rel 199+4 t=29 go.info."".sumAndMulWithNamedReturn$abstract+45
	rel 208+4 t=29 go.info."".sumAndMulWithNamedReturn$abstract+55
	rel 218+4 t=29 go.info."".sum$abstract+0
	rel 222+8 t=1 "".main+682
	rel 230+8 t=1 "".main+695
	rel 238+4 t=30 gofile../root/go/src/github.com/DQNEO/babygo/sample/sample.go+0
	rel 244+4 t=29 go.info."".sum$abstract+8
	rel 253+4 t=29 go.info."".sum$abstract+16
	rel 263+4 t=29 go.info."".concate$abstract+0
	rel 267+8 t=1 "".main+775
	rel 275+8 t=1 "".main+854
	rel 283+4 t=30 gofile../root/go/src/github.com/DQNEO/babygo/sample/sample.go+0
	rel 289+4 t=29 go.info."".concate$abstract+12
	rel 298+4 t=29 go.info."".concate$abstract+20
go.range."".main SDWARFRANGE size=64
	0x0000 ff ff ff ff ff ff ff ff 00 00 00 00 00 00 00 00  ................
	0x0010 ee 01 00 00 00 00 00 00 fb 01 00 00 00 00 00 00  ................
	0x0020 03 02 00 00 00 00 00 00 14 02 00 00 00 00 00 00  ................
	0x0030 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	rel 8+8 t=1 "".main+0
go.debuglines."".main SDWARFMISC size=153
	0x0000 04 02 03 28 14 0a 08 2d f6 24 06 87 06 08 03 63  ...(...-.$.....c
	0x0010 6f 06 5f 06 08 03 1c b4 06 5f 06 c4 06 55 06 02  o._......_...U..
	0x0020 29 ff 06 5f 06 02 37 03 62 fb 06 5f 06 08 03 1d  ).._..7.b.._....
	0x0030 b4 06 5f 06 c4 06 87 06 24 06 87 06 08 10 06 87  .._.....$.......
	0x0040 06 08 03 6c 29 06 5f 03 13 46 03 6c 5b 06 03 13  ...l)._..F.l[...
	0x0050 be 06 5f 06 08 38 06 69 06 08 03 6f 65 06 41 06  .._..8.i...oe.A.
	0x0060 92 06 41 06 03 0f be 24 06 87 06 03 75 bf 06 5f  ..A....$....u.._
	0x0070 06 03 0a 46 06 5f 06 56 06 55 06 02 22 03 78 fb  ...F._.V.U..".x.
	0x0080 06 5f 06 02 20 ff 06 41 06 03 07 78 06 5f 06 08  ._.. ..A...x._..
	0x0090 23 03 7b ab 04 01 03 53 01                       #.{....S.
""..inittask SNOPTRDATA size=24
	0x0000 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0010 00 00 00 00 00 00 00 00                          ........
runtime.memequal64·f SRODATA dupok size=8
	0x0000 00 00 00 00 00 00 00 00                          ........
	rel 0+8 t=1 runtime.memequal64+0
runtime.gcbits.01 SRODATA dupok size=1
	0x0000 01                                               .
type..namedata.*[]uint8- SRODATA dupok size=11
	0x0000 00 00 08 2a 5b 5d 75 69 6e 74 38                 ...*[]uint8
type.*[]uint8 SRODATA dupok size=56
	0x0000 08 00 00 00 00 00 00 00 08 00 00 00 00 00 00 00  ................
	0x0010 a5 8e d0 69 08 08 08 36 00 00 00 00 00 00 00 00  ...i...6........
	0x0020 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0030 00 00 00 00 00 00 00 00                          ........
	rel 24+8 t=1 runtime.memequal64·f+0
	rel 32+8 t=1 runtime.gcbits.01+0
	rel 40+4 t=5 type..namedata.*[]uint8-+0
	rel 48+8 t=1 type.[]uint8+0
type.[]uint8 SRODATA dupok size=56
	0x0000 18 00 00 00 00 00 00 00 08 00 00 00 00 00 00 00  ................
	0x0010 df 7e 2e 38 02 08 08 17 00 00 00 00 00 00 00 00  .~.8............
	0x0020 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0030 00 00 00 00 00 00 00 00                          ........
	rel 32+8 t=1 runtime.gcbits.01+0
	rel 40+4 t=5 type..namedata.*[]uint8-+0
	rel 44+4 t=6 type.*[]uint8+0
	rel 48+8 t=1 type.uint8+0
type..eqfunc3 SRODATA dupok size=16
	0x0000 00 00 00 00 00 00 00 00 03 00 00 00 00 00 00 00  ................
	rel 0+8 t=1 runtime.memequal_varlen+0
runtime.gcbits. SRODATA dupok size=0
type..namedata.*[3]uint8- SRODATA dupok size=12
	0x0000 00 00 09 2a 5b 33 5d 75 69 6e 74 38              ...*[3]uint8
type.[3]uint8 SRODATA dupok size=72
	0x0000 03 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0010 c2 b9 52 dd 0a 01 01 11 00 00 00 00 00 00 00 00  ..R.............
	0x0020 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0030 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0040 03 00 00 00 00 00 00 00                          ........
	rel 24+8 t=1 type..eqfunc3+0
	rel 32+8 t=1 runtime.gcbits.+0
	rel 40+4 t=5 type..namedata.*[3]uint8-+0
	rel 44+4 t=6 type.*[3]uint8+0
	rel 48+8 t=1 type.uint8+0
	rel 56+8 t=1 type.[]uint8+0
type.*[3]uint8 SRODATA dupok size=56
	0x0000 08 00 00 00 00 00 00 00 08 00 00 00 00 00 00 00  ................
	0x0010 69 04 66 6c 08 08 08 36 00 00 00 00 00 00 00 00  i.fl...6........
	0x0020 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0030 00 00 00 00 00 00 00 00                          ........
	rel 24+8 t=1 runtime.memequal64·f+0
	rel 32+8 t=1 runtime.gcbits.01+0
	rel 40+4 t=5 type..namedata.*[3]uint8-+0
	rel 48+8 t=1 type.[3]uint8+0
type..eqfunc32 SRODATA dupok size=16
	0x0000 00 00 00 00 00 00 00 00 20 00 00 00 00 00 00 00  ........ .......
	rel 0+8 t=1 runtime.memequal_varlen+0
type..namedata.*[32]uint8- SRODATA dupok size=13
	0x0000 00 00 0a 2a 5b 33 32 5d 75 69 6e 74 38           ...*[32]uint8
type.*[32]uint8 SRODATA dupok size=56
	0x0000 08 00 00 00 00 00 00 00 08 00 00 00 00 00 00 00  ................
	0x0010 f4 c7 79 15 08 08 08 36 00 00 00 00 00 00 00 00  ..y....6........
	0x0020 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0030 00 00 00 00 00 00 00 00                          ........
	rel 24+8 t=1 runtime.memequal64·f+0
	rel 32+8 t=1 runtime.gcbits.01+0
	rel 40+4 t=5 type..namedata.*[32]uint8-+0
	rel 48+8 t=1 type.[32]uint8+0
type.[32]uint8 SRODATA dupok size=72
	0x0000 20 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00   ...............
	0x0010 9c 59 ff a8 0a 01 01 11 00 00 00 00 00 00 00 00  .Y..............
	0x0020 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0030 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0040 20 00 00 00 00 00 00 00                           .......
	rel 24+8 t=1 type..eqfunc32+0
	rel 32+8 t=1 runtime.gcbits.+0
	rel 40+4 t=5 type..namedata.*[32]uint8-+0
	rel 44+4 t=6 type.*[32]uint8+0
	rel 48+8 t=1 type.uint8+0
	rel 56+8 t=1 type.[]uint8+0
gclocals·33cdeccccebe80329f1fdbee7f5874cb SRODATA dupok size=8
	0x0000 01 00 00 00 00 00 00 00                          ........
gclocals·305be4e7b0a18f2d72ac308e38ac4e7b SRODATA dupok size=11
	0x0000 03 00 00 00 04 00 00 00 01 00 08                 ...........
gclocals·cadea2e49003779a155f5f8fb1f0fe78 SRODATA dupok size=11
	0x0000 03 00 00 00 03 00 00 00 00 00 00                 ...........
gclocals·9fb7f0986f647f17cb53dda1484e0f7a SRODATA dupok size=10
	0x0000 02 00 00 00 01 00 00 00 00 01                    ..........
gclocals·2625d1fdbbaf79a2e52296235cb6527c SRODATA dupok size=12
	0x0000 04 00 00 00 05 00 00 00 05 04 00 10              ............
gclocals·f6bd6b3389b872033d462029172c8612 SRODATA dupok size=8
	0x0000 04 00 00 00 00 00 00 00                          ........
gclocals·1cf923758aae2e428391d1783fe59973 SRODATA dupok size=11
	0x0000 03 00 00 00 02 00 00 00 00 01 02                 ...........
gclocals·f5be5308b59e045b7c5b33ee8908cfb7 SRODATA dupok size=8
	0x0000 07 00 00 00 00 00 00 00                          ........
gclocals·a5f8d5878f569589345be9c78b1f4c93 SRODATA dupok size=36
	0x0000 07 00 00 00 1e 00 00 00 00 00 00 00 00 00 00 01  ................
	0x0010 01 00 00 00 00 00 20 00 20 00 00 00 28 00 00 00  ...... . ...(...
	0x0020 08 00 00 00                                      ....
