"".min STEXT nosplit size=1 args=0x0 locals=0x0
	0x0000 00000 (sample.go:5)	TEXT	"".min(SB), NOSPLIT|ABIInternal, $0-0
	0x0000 00000 (sample.go:5)	FUNCDATA	$0, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0000 00000 (sample.go:5)	FUNCDATA	$1, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0000 00000 (sample.go:6)	RET
	0x0000 c3                                               .
"".receiveBytes STEXT nosplit size=42 args=0x10 locals=0x10
	0x0000 00000 (sample.go:8)	TEXT	"".receiveBytes(SB), NOSPLIT|ABIInternal, $16-16
	0x0000 00000 (sample.go:8)	SUBQ	$16, SP
	0x0004 00004 (sample.go:8)	MOVQ	BP, 8(SP)
	0x0009 00009 (sample.go:8)	LEAQ	8(SP), BP
	0x000e 00014 (sample.go:8)	FUNCDATA	$0, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x000e 00014 (sample.go:8)	FUNCDATA	$1, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x000e 00014 (sample.go:8)	MOVB	$0, "".~r3+32(SP)
	0x0013 00019 (sample.go:9)	MOVBLZX	"".a+24(SP), AX
	0x0018 00024 (sample.go:9)	MOVB	AL, "".r+7(SP)
	0x001c 00028 (sample.go:10)	MOVB	AL, "".~r3+32(SP)
	0x0020 00032 (sample.go:10)	MOVQ	8(SP), BP
	0x0025 00037 (sample.go:10)	ADDQ	$16, SP
	0x0029 00041 (sample.go:10)	RET
	0x0000 48 83 ec 10 48 89 6c 24 08 48 8d 6c 24 08 c6 44  H...H.l$.H.l$..D
	0x0010 24 20 00 0f b6 44 24 18 88 44 24 07 88 44 24 20  $ ...D$..D$..D$ 
	0x0020 48 8b 6c 24 08 48 83 c4 10 c3                    H.l$.H....
"".testPassBytes STEXT size=133 args=0x0 locals=0x20
	0x0000 00000 (sample.go:13)	TEXT	"".testPassBytes(SB), ABIInternal, $32-0
	0x0000 00000 (sample.go:13)	MOVQ	(TLS), CX
	0x0009 00009 (sample.go:13)	CMPQ	SP, 16(CX)
	0x000d 00013 (sample.go:13)	PCDATA	$0, $-2
	0x000d 00013 (sample.go:13)	JLS	121
	0x000f 00015 (sample.go:13)	PCDATA	$0, $-1
	0x000f 00015 (sample.go:13)	SUBQ	$32, SP
	0x0013 00019 (sample.go:13)	MOVQ	BP, 24(SP)
	0x0018 00024 (sample.go:13)	LEAQ	24(SP), BP
	0x001d 00029 (sample.go:13)	FUNCDATA	$0, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x001d 00029 (sample.go:13)	FUNCDATA	$1, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x001d 00029 (sample.go:14)	MOVB	$97, "".a+23(SP)
	0x0022 00034 (sample.go:15)	MOVB	$98, "".b+22(SP)
	0x0027 00039 (sample.go:16)	MOVB	$99, "".c+21(SP)
	0x002c 00044 (sample.go:17)	MOVBLZX	"".a+23(SP), AX
	0x0031 00049 (sample.go:17)	MOVB	AL, (SP)
	0x0034 00052 (sample.go:17)	MOVBLZX	"".b+22(SP), AX
	0x0039 00057 (sample.go:17)	MOVB	AL, 1(SP)
	0x003d 00061 (sample.go:17)	MOVB	$99, 2(SP)
	0x0042 00066 (sample.go:17)	PCDATA	$1, $0
	0x0042 00066 (sample.go:17)	CALL	"".receiveBytes(SB)
	0x0047 00071 (sample.go:17)	MOVBLZX	8(SP), AX
	0x004c 00076 (sample.go:17)	MOVB	AL, "".d+20(SP)
	0x0050 00080 (sample.go:18)	CALL	runtime.printlock(SB)
	0x0055 00085 (sample.go:18)	MOVBLZX	"".d+20(SP), AX
	0x005a 00090 (sample.go:18)	MOVQ	AX, (SP)
	0x005e 00094 (sample.go:18)	NOP
	0x0060 00096 (sample.go:18)	CALL	runtime.printuint(SB)
	0x0065 00101 (sample.go:18)	CALL	runtime.printnl(SB)
	0x006a 00106 (sample.go:18)	CALL	runtime.printunlock(SB)
	0x006f 00111 (sample.go:19)	MOVQ	24(SP), BP
	0x0074 00116 (sample.go:19)	ADDQ	$32, SP
	0x0078 00120 (sample.go:19)	RET
	0x0079 00121 (sample.go:19)	NOP
	0x0079 00121 (sample.go:13)	PCDATA	$1, $-1
	0x0079 00121 (sample.go:13)	PCDATA	$0, $-2
	0x0079 00121 (sample.go:13)	CALL	runtime.morestack_noctxt(SB)
	0x007e 00126 (sample.go:13)	PCDATA	$0, $-1
	0x007e 00126 (sample.go:13)	NOP
	0x0080 00128 (sample.go:13)	JMP	0
	0x0000 64 48 8b 0c 25 00 00 00 00 48 3b 61 10 76 6a 48  dH..%....H;a.vjH
	0x0010 83 ec 20 48 89 6c 24 18 48 8d 6c 24 18 c6 44 24  .. H.l$.H.l$..D$
	0x0020 17 61 c6 44 24 16 62 c6 44 24 15 63 0f b6 44 24  .a.D$.b.D$.c..D$
	0x0030 17 88 04 24 0f b6 44 24 16 88 44 24 01 c6 44 24  ...$..D$..D$..D$
	0x0040 02 63 e8 00 00 00 00 0f b6 44 24 08 88 44 24 14  .c.......D$..D$.
	0x0050 e8 00 00 00 00 0f b6 44 24 14 48 89 04 24 66 90  .......D$.H..$f.
	0x0060 e8 00 00 00 00 e8 00 00 00 00 e8 00 00 00 00 48  ...............H
	0x0070 8b 6c 24 18 48 83 c4 20 c3 e8 00 00 00 00 66 90  .l$.H.. ......f.
	0x0080 e9 7b ff ff ff                                   .{...
	rel 5+4 t=17 TLS+0
	rel 67+4 t=8 "".receiveBytes+0
	rel 81+4 t=8 runtime.printlock+0
	rel 97+4 t=8 runtime.printuint+0
	rel 102+4 t=8 runtime.printnl+0
	rel 107+4 t=8 runtime.printunlock+0
	rel 122+4 t=8 runtime.morestack_noctxt+0
"".char STEXT nosplit size=48 args=0x10 locals=0x10
	0x0000 00000 (sample.go:21)	TEXT	"".char(SB), NOSPLIT|ABIInternal, $16-16
	0x0000 00000 (sample.go:21)	SUBQ	$16, SP
	0x0004 00004 (sample.go:21)	MOVQ	BP, 8(SP)
	0x0009 00009 (sample.go:21)	LEAQ	8(SP), BP
	0x000e 00014 (sample.go:21)	FUNCDATA	$0, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x000e 00014 (sample.go:21)	FUNCDATA	$1, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x000e 00014 (sample.go:21)	MOVB	$0, "".~r1+32(SP)
	0x0013 00019 (sample.go:22)	MOVBLZX	"".a+24(SP), AX
	0x0018 00024 (sample.go:22)	MOVB	AL, "".b+7(SP)
	0x001c 00028 (sample.go:23)	MOVB	$65, "".b+7(SP)
	0x0021 00033 (sample.go:24)	MOVB	$65, "".~r1+32(SP)
	0x0026 00038 (sample.go:24)	MOVQ	8(SP), BP
	0x002b 00043 (sample.go:24)	ADDQ	$16, SP
	0x002f 00047 (sample.go:24)	RET
	0x0000 48 83 ec 10 48 89 6c 24 08 48 8d 6c 24 08 c6 44  H...H.l$.H.l$..D
	0x0010 24 20 00 0f b6 44 24 18 88 44 24 07 c6 44 24 07  $ ...D$..D$..D$.
	0x0020 41 c6 44 24 20 41 48 8b 6c 24 08 48 83 c4 10 c3  A.D$ AH.l$.H....
"".slice STEXT nosplit size=85 args=0x30 locals=0x20
	0x0000 00000 (sample.go:27)	TEXT	"".slice(SB), NOSPLIT|ABIInternal, $32-48
	0x0000 00000 (sample.go:27)	SUBQ	$32, SP
	0x0004 00004 (sample.go:27)	MOVQ	BP, 24(SP)
	0x0009 00009 (sample.go:27)	LEAQ	24(SP), BP
	0x000e 00014 (sample.go:27)	FUNCDATA	$0, gclocals·4032f753396f2012ad1784f398b170f4(SB)
	0x000e 00014 (sample.go:27)	FUNCDATA	$1, gclocals·15b76348caca8a511afecadf603e9401(SB)
	0x000e 00014 (sample.go:27)	MOVQ	$0, "".~r1+64(SP)
	0x0017 00023 (sample.go:27)	XORPS	X0, X0
	0x001a 00026 (sample.go:27)	MOVUPS	X0, "".~r1+72(SP)
	0x001f 00031 (sample.go:28)	MOVQ	"".a+40(SP), AX
	0x0024 00036 (sample.go:28)	MOVQ	"".a+48(SP), CX
	0x0029 00041 (sample.go:28)	MOVQ	"".a+56(SP), DX
	0x002e 00046 (sample.go:28)	MOVQ	AX, "".b(SP)
	0x0032 00050 (sample.go:28)	MOVQ	CX, "".b+8(SP)
	0x0037 00055 (sample.go:28)	MOVQ	DX, "".b+16(SP)
	0x003c 00060 (sample.go:29)	MOVQ	AX, "".~r1+64(SP)
	0x0041 00065 (sample.go:29)	MOVQ	CX, "".~r1+72(SP)
	0x0046 00070 (sample.go:29)	MOVQ	DX, "".~r1+80(SP)
	0x004b 00075 (sample.go:29)	MOVQ	24(SP), BP
	0x0050 00080 (sample.go:29)	ADDQ	$32, SP
	0x0054 00084 (sample.go:29)	RET
	0x0000 48 83 ec 20 48 89 6c 24 18 48 8d 6c 24 18 48 c7  H.. H.l$.H.l$.H.
	0x0010 44 24 40 00 00 00 00 0f 57 c0 0f 11 44 24 48 48  D$@.....W...D$HH
	0x0020 8b 44 24 28 48 8b 4c 24 30 48 8b 54 24 38 48 89  .D$(H.L$0H.T$8H.
	0x0030 04 24 48 89 4c 24 08 48 89 54 24 10 48 89 44 24  .$H.L$.H.T$.H.D$
	0x0040 40 48 89 4c 24 48 48 89 54 24 50 48 8b 6c 24 18  @H.L$HH.T$PH.l$.
	0x0050 48 83 c4 20 c3                                   H.. .
"".arg1 STEXT nosplit size=1 args=0x8 locals=0x0
	0x0000 00000 (sample.go:32)	TEXT	"".arg1(SB), NOSPLIT|ABIInternal, $0-8
	0x0000 00000 (sample.go:32)	FUNCDATA	$0, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0000 00000 (sample.go:32)	FUNCDATA	$1, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0000 00000 (sample.go:34)	RET
	0x0000 c3                                               .
"".arg1ret1 STEXT nosplit size=20 args=0x10 locals=0x0
	0x0000 00000 (sample.go:36)	TEXT	"".arg1ret1(SB), NOSPLIT|ABIInternal, $0-16
	0x0000 00000 (sample.go:36)	FUNCDATA	$0, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0000 00000 (sample.go:36)	FUNCDATA	$1, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0000 00000 (sample.go:36)	MOVQ	$0, "".~r1+16(SP)
	0x0009 00009 (sample.go:37)	MOVQ	"".x+8(SP), AX
	0x000e 00014 (sample.go:37)	MOVQ	AX, "".~r1+16(SP)
	0x0013 00019 (sample.go:37)	RET
	0x0000 48 c7 44 24 10 00 00 00 00 48 8b 44 24 08 48 89  H.D$.....H.D$.H.
	0x0010 44 24 10 c3                                      D$..
"".sumAndMul STEXT nosplit size=53 args=0x20 locals=0x0
	0x0000 00000 (sample.go:40)	TEXT	"".sumAndMul(SB), NOSPLIT|ABIInternal, $0-32
	0x0000 00000 (sample.go:40)	FUNCDATA	$0, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0000 00000 (sample.go:40)	FUNCDATA	$1, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0000 00000 (sample.go:40)	MOVQ	$0, "".~r2+24(SP)
	0x0009 00009 (sample.go:40)	MOVQ	$0, "".~r3+32(SP)
	0x0012 00018 (sample.go:41)	MOVQ	"".x+8(SP), AX
	0x0017 00023 (sample.go:41)	ADDQ	"".y+16(SP), AX
	0x001c 00028 (sample.go:41)	MOVQ	AX, "".~r2+24(SP)
	0x0021 00033 (sample.go:41)	MOVQ	"".x+8(SP), AX
	0x0026 00038 (sample.go:41)	MOVQ	"".y+16(SP), CX
	0x002b 00043 (sample.go:41)	IMULQ	CX, AX
	0x002f 00047 (sample.go:41)	MOVQ	AX, "".~r3+32(SP)
	0x0034 00052 (sample.go:41)	RET
	0x0000 48 c7 44 24 18 00 00 00 00 48 c7 44 24 20 00 00  H.D$.....H.D$ ..
	0x0010 00 00 48 8b 44 24 08 48 03 44 24 10 48 89 44 24  ..H.D$.H.D$.H.D$
	0x0020 18 48 8b 44 24 08 48 8b 4c 24 10 48 0f af c1 48  .H.D$.H.L$.H...H
	0x0030 89 44 24 20 c3                                   .D$ .
"".sumAndMulWithNamedReturn STEXT nosplit size=53 args=0x20 locals=0x0
	0x0000 00000 (sample.go:44)	TEXT	"".sumAndMulWithNamedReturn(SB), NOSPLIT|ABIInternal, $0-32
	0x0000 00000 (sample.go:44)	FUNCDATA	$0, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0000 00000 (sample.go:44)	FUNCDATA	$1, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0000 00000 (sample.go:44)	MOVQ	$0, "".sum+24(SP)
	0x0009 00009 (sample.go:44)	MOVQ	$0, "".mul+32(SP)
	0x0012 00018 (sample.go:45)	MOVQ	"".x+8(SP), AX
	0x0017 00023 (sample.go:45)	ADDQ	"".y+16(SP), AX
	0x001c 00028 (sample.go:45)	MOVQ	AX, "".sum+24(SP)
	0x0021 00033 (sample.go:46)	MOVQ	"".y+16(SP), AX
	0x0026 00038 (sample.go:46)	MOVQ	"".x+8(SP), CX
	0x002b 00043 (sample.go:46)	IMULQ	AX, CX
	0x002f 00047 (sample.go:46)	MOVQ	CX, "".mul+32(SP)
	0x0034 00052 (sample.go:47)	RET
	0x0000 48 c7 44 24 18 00 00 00 00 48 c7 44 24 20 00 00  H.D$.....H.D$ ..
	0x0010 00 00 48 8b 44 24 08 48 03 44 24 10 48 89 44 24  ..H.D$.H.D$.H.D$
	0x0020 18 48 8b 44 24 10 48 8b 4c 24 08 48 0f af c8 48  .H.D$.H.L$.H...H
	0x0030 89 4c 24 20 c3                                   .L$ .
"".sum STEXT nosplit size=25 args=0x18 locals=0x0
	0x0000 00000 (sample.go:51)	TEXT	"".sum(SB), NOSPLIT|ABIInternal, $0-24
	0x0000 00000 (sample.go:51)	FUNCDATA	$0, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0000 00000 (sample.go:51)	FUNCDATA	$1, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0000 00000 (sample.go:51)	MOVQ	$0, "".~r2+24(SP)
	0x0009 00009 (sample.go:52)	MOVQ	"".x+8(SP), AX
	0x000e 00014 (sample.go:52)	ADDQ	"".y+16(SP), AX
	0x0013 00019 (sample.go:52)	MOVQ	AX, "".~r2+24(SP)
	0x0018 00024 (sample.go:52)	RET
	0x0000 48 c7 44 24 18 00 00 00 00 48 8b 44 24 08 48 03  H.D$.....H.D$.H.
	0x0010 44 24 10 48 89 44 24 18 c3                       D$.H.D$..
"".concate STEXT size=127 args=0x30 locals=0x40
	0x0000 00000 (sample.go:55)	TEXT	"".concate(SB), ABIInternal, $64-48
	0x0000 00000 (sample.go:55)	MOVQ	(TLS), CX
	0x0009 00009 (sample.go:55)	CMPQ	SP, 16(CX)
	0x000d 00013 (sample.go:55)	PCDATA	$0, $-2
	0x000d 00013 (sample.go:55)	JLS	120
	0x000f 00015 (sample.go:55)	PCDATA	$0, $-1
	0x000f 00015 (sample.go:55)	SUBQ	$64, SP
	0x0013 00019 (sample.go:55)	MOVQ	BP, 56(SP)
	0x0018 00024 (sample.go:55)	LEAQ	56(SP), BP
	0x001d 00029 (sample.go:55)	FUNCDATA	$0, gclocals·5207c493e17be99b5ba2331b72d2d660(SB)
	0x001d 00029 (sample.go:55)	FUNCDATA	$1, gclocals·69c1753bd5f81501d95132d08af04464(SB)
	0x001d 00029 (sample.go:55)	XORPS	X0, X0
	0x0020 00032 (sample.go:55)	MOVUPS	X0, "".~r2+104(SP)
	0x0025 00037 (sample.go:56)	MOVQ	$0, (SP)
	0x002d 00045 (sample.go:56)	MOVQ	"".x+72(SP), AX
	0x0032 00050 (sample.go:56)	MOVQ	"".x+80(SP), CX
	0x0037 00055 (sample.go:56)	MOVQ	AX, 8(SP)
	0x003c 00060 (sample.go:56)	MOVQ	CX, 16(SP)
	0x0041 00065 (sample.go:56)	MOVQ	"".y+88(SP), AX
	0x0046 00070 (sample.go:56)	MOVQ	"".y+96(SP), CX
	0x004b 00075 (sample.go:56)	MOVQ	AX, 24(SP)
	0x0050 00080 (sample.go:56)	MOVQ	CX, 32(SP)
	0x0055 00085 (sample.go:56)	PCDATA	$1, $1
	0x0055 00085 (sample.go:56)	CALL	runtime.concatstring2(SB)
	0x005a 00090 (sample.go:56)	MOVQ	48(SP), AX
	0x005f 00095 (sample.go:56)	MOVQ	40(SP), CX
	0x0064 00100 (sample.go:56)	MOVQ	CX, "".~r2+104(SP)
	0x0069 00105 (sample.go:56)	MOVQ	AX, "".~r2+112(SP)
	0x006e 00110 (sample.go:56)	MOVQ	56(SP), BP
	0x0073 00115 (sample.go:56)	ADDQ	$64, SP
	0x0077 00119 (sample.go:56)	RET
	0x0078 00120 (sample.go:56)	NOP
	0x0078 00120 (sample.go:55)	PCDATA	$1, $-1
	0x0078 00120 (sample.go:55)	PCDATA	$0, $-2
	0x0078 00120 (sample.go:55)	CALL	runtime.morestack_noctxt(SB)
	0x007d 00125 (sample.go:55)	PCDATA	$0, $-1
	0x007d 00125 (sample.go:55)	JMP	0
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
"".main STEXT size=330 args=0x0 locals=0x60
	0x0000 00000 (sample.go:59)	TEXT	"".main(SB), ABIInternal, $96-0
	0x0000 00000 (sample.go:59)	MOVQ	(TLS), CX
	0x0009 00009 (sample.go:59)	CMPQ	SP, 16(CX)
	0x000d 00013 (sample.go:59)	PCDATA	$0, $-2
	0x000d 00013 (sample.go:59)	JLS	320
	0x0013 00019 (sample.go:59)	PCDATA	$0, $-1
	0x0013 00019 (sample.go:59)	SUBQ	$96, SP
	0x0017 00023 (sample.go:59)	MOVQ	BP, 88(SP)
	0x001c 00028 (sample.go:59)	LEAQ	88(SP), BP
	0x0021 00033 (sample.go:59)	FUNCDATA	$0, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0021 00033 (sample.go:59)	FUNCDATA	$1, gclocals·ff19ed39bdde8a01a800918ac3ef0ec7(SB)
	0x0021 00033 (sample.go:60)	PCDATA	$1, $0
	0x0021 00033 (sample.go:60)	CALL	"".min(SB)
	0x0026 00038 (sample.go:61)	MOVQ	$0, (SP)
	0x002e 00046 (sample.go:61)	XORPS	X0, X0
	0x0031 00049 (sample.go:61)	MOVUPS	X0, 8(SP)
	0x0036 00054 (sample.go:61)	CALL	"".slice(SB)
	0x003b 00059 (sample.go:62)	MOVW	$0, ""..autotmp_0+53(SP)
	0x0042 00066 (sample.go:62)	MOVB	$0, ""..autotmp_0+55(SP)
	0x0047 00071 (sample.go:62)	LEAQ	""..autotmp_0+53(SP), AX
	0x004c 00076 (sample.go:62)	MOVQ	AX, ""..autotmp_2+56(SP)
	0x0051 00081 (sample.go:62)	TESTB	AL, (AX)
	0x0053 00083 (sample.go:62)	MOVB	$97, ""..autotmp_0+53(SP)
	0x0058 00088 (sample.go:62)	MOVQ	""..autotmp_2+56(SP), AX
	0x005d 00093 (sample.go:62)	TESTB	AL, (AX)
	0x005f 00095 (sample.go:62)	MOVB	$98, 1(AX)
	0x0063 00099 (sample.go:62)	MOVQ	""..autotmp_2+56(SP), AX
	0x0068 00104 (sample.go:62)	TESTB	AL, (AX)
	0x006a 00106 (sample.go:62)	MOVB	$99, 2(AX)
	0x006e 00110 (sample.go:62)	MOVQ	""..autotmp_2+56(SP), AX
	0x0073 00115 (sample.go:62)	TESTB	AL, (AX)
	0x0075 00117 (sample.go:62)	JMP	119
	0x0077 00119 (sample.go:62)	MOVQ	AX, ""..autotmp_1+64(SP)
	0x007c 00124 (sample.go:62)	MOVQ	$3, ""..autotmp_1+72(SP)
	0x0085 00133 (sample.go:62)	MOVQ	$3, ""..autotmp_1+80(SP)
	0x008e 00142 (sample.go:62)	MOVQ	AX, (SP)
	0x0092 00146 (sample.go:62)	MOVQ	$3, 8(SP)
	0x009b 00155 (sample.go:62)	MOVQ	$3, 16(SP)
	0x00a4 00164 (sample.go:62)	CALL	"".slice(SB)
	0x00a9 00169 (sample.go:63)	MOVQ	$1, (SP)
	0x00b1 00177 (sample.go:63)	CALL	"".arg1(SB)
	0x00b6 00182 (sample.go:64)	MOVQ	$2, (SP)
	0x00be 00190 (sample.go:64)	NOP
	0x00c0 00192 (sample.go:64)	CALL	"".arg1ret1(SB)
	0x00c5 00197 (sample.go:65)	MOVQ	$5, (SP)
	0x00cd 00205 (sample.go:65)	MOVQ	$7, 8(SP)
	0x00d6 00214 (sample.go:65)	CALL	"".sumAndMul(SB)
	0x00db 00219 (sample.go:66)	MOVQ	$5, (SP)
	0x00e3 00227 (sample.go:66)	MOVQ	$7, 8(SP)
	0x00ec 00236 (sample.go:66)	CALL	"".sumAndMulWithNamedReturn(SB)
	0x00f1 00241 (sample.go:67)	MOVQ	$2, (SP)
	0x00f9 00249 (sample.go:67)	MOVQ	$3, 8(SP)
	0x0102 00258 (sample.go:67)	CALL	"".sum(SB)
	0x0107 00263 (sample.go:68)	LEAQ	go.string."hello"(SB), AX
	0x010e 00270 (sample.go:68)	MOVQ	AX, (SP)
	0x0112 00274 (sample.go:68)	MOVQ	$5, 8(SP)
	0x011b 00283 (sample.go:68)	LEAQ	go.string." world"(SB), AX
	0x0122 00290 (sample.go:68)	MOVQ	AX, 16(SP)
	0x0127 00295 (sample.go:68)	MOVQ	$6, 24(SP)
	0x0130 00304 (sample.go:68)	CALL	"".concate(SB)
	0x0135 00309 (sample.go:69)	MOVQ	88(SP), BP
	0x013a 00314 (sample.go:69)	ADDQ	$96, SP
	0x013e 00318 (sample.go:69)	RET
	0x013f 00319 (sample.go:69)	NOP
	0x013f 00319 (sample.go:59)	PCDATA	$1, $-1
	0x013f 00319 (sample.go:59)	PCDATA	$0, $-2
	0x013f 00319 (sample.go:59)	NOP
	0x0140 00320 (sample.go:59)	CALL	runtime.morestack_noctxt(SB)
	0x0145 00325 (sample.go:59)	PCDATA	$0, $-1
	0x0145 00325 (sample.go:59)	JMP	0
	0x0000 64 48 8b 0c 25 00 00 00 00 48 3b 61 10 0f 86 2d  dH..%....H;a...-
	0x0010 01 00 00 48 83 ec 60 48 89 6c 24 58 48 8d 6c 24  ...H..`H.l$XH.l$
	0x0020 58 e8 00 00 00 00 48 c7 04 24 00 00 00 00 0f 57  X.....H..$.....W
	0x0030 c0 0f 11 44 24 08 e8 00 00 00 00 66 c7 44 24 35  ...D$......f.D$5
	0x0040 00 00 c6 44 24 37 00 48 8d 44 24 35 48 89 44 24  ...D$7.H.D$5H.D$
	0x0050 38 84 00 c6 44 24 35 61 48 8b 44 24 38 84 00 c6  8...D$5aH.D$8...
	0x0060 40 01 62 48 8b 44 24 38 84 00 c6 40 02 63 48 8b  @.bH.D$8...@.cH.
	0x0070 44 24 38 84 00 eb 00 48 89 44 24 40 48 c7 44 24  D$8....H.D$@H.D$
	0x0080 48 03 00 00 00 48 c7 44 24 50 03 00 00 00 48 89  H....H.D$P....H.
	0x0090 04 24 48 c7 44 24 08 03 00 00 00 48 c7 44 24 10  .$H.D$.....H.D$.
	0x00a0 03 00 00 00 e8 00 00 00 00 48 c7 04 24 01 00 00  .........H..$...
	0x00b0 00 e8 00 00 00 00 48 c7 04 24 02 00 00 00 66 90  ......H..$....f.
	0x00c0 e8 00 00 00 00 48 c7 04 24 05 00 00 00 48 c7 44  .....H..$....H.D
	0x00d0 24 08 07 00 00 00 e8 00 00 00 00 48 c7 04 24 05  $..........H..$.
	0x00e0 00 00 00 48 c7 44 24 08 07 00 00 00 e8 00 00 00  ...H.D$.........
	0x00f0 00 48 c7 04 24 02 00 00 00 48 c7 44 24 08 03 00  .H..$....H.D$...
	0x0100 00 00 e8 00 00 00 00 48 8d 05 00 00 00 00 48 89  .......H......H.
	0x0110 04 24 48 c7 44 24 08 05 00 00 00 48 8d 05 00 00  .$H.D$.....H....
	0x0120 00 00 48 89 44 24 10 48 c7 44 24 18 06 00 00 00  ..H.D$.H.D$.....
	0x0130 e8 00 00 00 00 48 8b 6c 24 58 48 83 c4 60 c3 90  .....H.l$XH..`..
	0x0140 e8 00 00 00 00 e9 b6 fe ff ff                    ..........
	rel 5+4 t=17 TLS+0
	rel 34+4 t=8 "".min+0
	rel 55+4 t=8 "".slice+0
	rel 165+4 t=8 "".slice+0
	rel 178+4 t=8 "".arg1+0
	rel 193+4 t=8 "".arg1ret1+0
	rel 215+4 t=8 "".sumAndMul+0
	rel 237+4 t=8 "".sumAndMulWithNamedReturn+0
	rel 259+4 t=8 "".sum+0
	rel 266+4 t=16 go.string."hello"+0
	rel 286+4 t=16 go.string." world"+0
	rel 305+4 t=8 "".concate+0
	rel 321+4 t=8 runtime.morestack_noctxt+0
go.cuinfo.packagename. SDWARFINFO dupok size=0
	0x0000 6d 61 69 6e                                      main
go.string."hello" SRODATA dupok size=5
	0x0000 68 65 6c 6c 6f                                   hello
go.string." world" SRODATA dupok size=6
	0x0000 20 77 6f 72 6c 64                                 world
""..inittask SNOPTRDATA size=24
	0x0000 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0010 00 00 00 00 00 00 00 00                          ........
gclocals·33cdeccccebe80329f1fdbee7f5874cb SRODATA dupok size=8
	0x0000 01 00 00 00 00 00 00 00                          ........
gclocals·4032f753396f2012ad1784f398b170f4 SRODATA dupok size=10
	0x0000 02 00 00 00 04 00 00 00 01 00                    ..........
gclocals·15b76348caca8a511afecadf603e9401 SRODATA dupok size=10
	0x0000 02 00 00 00 03 00 00 00 00 00                    ..........
gclocals·5207c493e17be99b5ba2331b72d2d660 SRODATA dupok size=10
	0x0000 02 00 00 00 05 00 00 00 05 00                    ..........
gclocals·69c1753bd5f81501d95132d08af04464 SRODATA dupok size=8
	0x0000 02 00 00 00 00 00 00 00                          ........
gclocals·ff19ed39bdde8a01a800918ac3ef0ec7 SRODATA dupok size=9
	0x0000 01 00 00 00 04 00 00 00 00                       .........
