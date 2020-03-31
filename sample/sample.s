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
"".sum STEXT nosplit size=25 args=0x18 locals=0x0
	0x0000 00000 (sample/sample.go:27)	TEXT	"".sum(SB), NOSPLIT|ABIInternal, $0-24
	0x0000 00000 (sample/sample.go:27)	PCDATA	$0, $-2
	0x0000 00000 (sample/sample.go:27)	PCDATA	$1, $-2
	0x0000 00000 (sample/sample.go:27)	FUNCDATA	$0, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0000 00000 (sample/sample.go:27)	FUNCDATA	$1, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0000 00000 (sample/sample.go:27)	FUNCDATA	$2, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0000 00000 (sample/sample.go:27)	PCDATA	$0, $0
	0x0000 00000 (sample/sample.go:27)	PCDATA	$1, $0
	0x0000 00000 (sample/sample.go:27)	MOVQ	$0, "".~r2+24(SP)
	0x0009 00009 (sample/sample.go:28)	MOVQ	"".x+8(SP), AX
	0x000e 00014 (sample/sample.go:28)	ADDQ	"".y+16(SP), AX
	0x0013 00019 (sample/sample.go:28)	MOVQ	AX, "".~r2+24(SP)
	0x0018 00024 (sample/sample.go:28)	RET
	0x0000 48 c7 44 24 18 00 00 00 00 48 8b 44 24 08 48 03  H.D$.....H.D$.H.
	0x0010 44 24 10 48 89 44 24 18 c3                       D$.H.D$..
"".concate STEXT size=127 args=0x30 locals=0x40
	0x0000 00000 (sample/sample.go:31)	TEXT	"".concate(SB), ABIInternal, $64-48
	0x0000 00000 (sample/sample.go:31)	MOVQ	(TLS), CX
	0x0009 00009 (sample/sample.go:31)	CMPQ	SP, 16(CX)
	0x000d 00013 (sample/sample.go:31)	PCDATA	$0, $-2
	0x000d 00013 (sample/sample.go:31)	JLS	120
	0x000f 00015 (sample/sample.go:31)	PCDATA	$0, $-1
	0x000f 00015 (sample/sample.go:31)	SUBQ	$64, SP
	0x0013 00019 (sample/sample.go:31)	MOVQ	BP, 56(SP)
	0x0018 00024 (sample/sample.go:31)	LEAQ	56(SP), BP
	0x001d 00029 (sample/sample.go:31)	PCDATA	$0, $-2
	0x001d 00029 (sample/sample.go:31)	PCDATA	$1, $-2
	0x001d 00029 (sample/sample.go:31)	FUNCDATA	$0, gclocals·2625d1fdbbaf79a2e52296235cb6527c(SB)
	0x001d 00029 (sample/sample.go:31)	FUNCDATA	$1, gclocals·f6bd6b3389b872033d462029172c8612(SB)
	0x001d 00029 (sample/sample.go:31)	FUNCDATA	$2, gclocals·1cf923758aae2e428391d1783fe59973(SB)
	0x001d 00029 (sample/sample.go:31)	PCDATA	$0, $0
	0x001d 00029 (sample/sample.go:31)	PCDATA	$1, $0
	0x001d 00029 (sample/sample.go:31)	XORPS	X0, X0
	0x0020 00032 (sample/sample.go:31)	MOVUPS	X0, "".~r2+104(SP)
	0x0025 00037 (sample/sample.go:32)	MOVQ	$0, (SP)
	0x002d 00045 (sample/sample.go:32)	PCDATA	$0, $1
	0x002d 00045 (sample/sample.go:32)	MOVQ	"".x+72(SP), AX
	0x0032 00050 (sample/sample.go:32)	PCDATA	$1, $1
	0x0032 00050 (sample/sample.go:32)	MOVQ	"".x+80(SP), CX
	0x0037 00055 (sample/sample.go:32)	PCDATA	$0, $0
	0x0037 00055 (sample/sample.go:32)	MOVQ	AX, 8(SP)
	0x003c 00060 (sample/sample.go:32)	MOVQ	CX, 16(SP)
	0x0041 00065 (sample/sample.go:32)	PCDATA	$0, $1
	0x0041 00065 (sample/sample.go:32)	MOVQ	"".y+88(SP), AX
	0x0046 00070 (sample/sample.go:32)	PCDATA	$1, $2
	0x0046 00070 (sample/sample.go:32)	MOVQ	"".y+96(SP), CX
	0x004b 00075 (sample/sample.go:32)	PCDATA	$0, $0
	0x004b 00075 (sample/sample.go:32)	MOVQ	AX, 24(SP)
	0x0050 00080 (sample/sample.go:32)	MOVQ	CX, 32(SP)
	0x0055 00085 (sample/sample.go:32)	CALL	runtime.concatstring2(SB)
	0x005a 00090 (sample/sample.go:32)	MOVQ	48(SP), AX
	0x005f 00095 (sample/sample.go:32)	PCDATA	$0, $2
	0x005f 00095 (sample/sample.go:32)	MOVQ	40(SP), CX
	0x0064 00100 (sample/sample.go:32)	PCDATA	$0, $0
	0x0064 00100 (sample/sample.go:32)	PCDATA	$1, $3
	0x0064 00100 (sample/sample.go:32)	MOVQ	CX, "".~r2+104(SP)
	0x0069 00105 (sample/sample.go:32)	MOVQ	AX, "".~r2+112(SP)
	0x006e 00110 (sample/sample.go:32)	MOVQ	56(SP), BP
	0x0073 00115 (sample/sample.go:32)	ADDQ	$64, SP
	0x0077 00119 (sample/sample.go:32)	RET
	0x0078 00120 (sample/sample.go:32)	NOP
	0x0078 00120 (sample/sample.go:31)	PCDATA	$1, $-1
	0x0078 00120 (sample/sample.go:31)	PCDATA	$0, $-2
	0x0078 00120 (sample/sample.go:31)	CALL	runtime.morestack_noctxt(SB)
	0x007d 00125 (sample/sample.go:31)	PCDATA	$0, $-1
	0x007d 00125 (sample/sample.go:31)	JMP	0
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
"".main STEXT size=696 args=0x0 locals=0x190
	0x0000 00000 (sample/sample.go:35)	TEXT	"".main(SB), ABIInternal, $400-0
	0x0000 00000 (sample/sample.go:35)	MOVQ	(TLS), CX
	0x0009 00009 (sample/sample.go:35)	LEAQ	-272(SP), AX
	0x0011 00017 (sample/sample.go:35)	CMPQ	AX, 16(CX)
	0x0015 00021 (sample/sample.go:35)	PCDATA	$0, $-2
	0x0015 00021 (sample/sample.go:35)	JLS	686
	0x001b 00027 (sample/sample.go:35)	PCDATA	$0, $-1
	0x001b 00027 (sample/sample.go:35)	SUBQ	$400, SP
	0x0022 00034 (sample/sample.go:35)	MOVQ	BP, 392(SP)
	0x002a 00042 (sample/sample.go:35)	LEAQ	392(SP), BP
	0x0032 00050 (sample/sample.go:35)	PCDATA	$0, $-2
	0x0032 00050 (sample/sample.go:35)	PCDATA	$1, $-2
	0x0032 00050 (sample/sample.go:35)	FUNCDATA	$0, gclocals·f5be5308b59e045b7c5b33ee8908cfb7(SB)
	0x0032 00050 (sample/sample.go:35)	FUNCDATA	$1, gclocals·ad9a9b86035be5675884f71ffff3e114(SB)
	0x0032 00050 (sample/sample.go:35)	FUNCDATA	$2, gclocals·9fb7f0986f647f17cb53dda1484e0f7a(SB)
	0x0032 00050 (sample/sample.go:36)	PCDATA	$0, $-1
	0x0032 00050 (sample/sample.go:36)	PCDATA	$1, $-1
	0x0032 00050 (sample/sample.go:36)	JMP	52
	0x0034 00052 (sample/sample.go:37)	PCDATA	$0, $0
	0x0034 00052 (sample/sample.go:37)	PCDATA	$1, $1
	0x0034 00052 (sample/sample.go:37)	MOVQ	$0, "".a+320(SP)
	0x0040 00064 (sample/sample.go:37)	XORPS	X0, X0
	0x0043 00067 (sample/sample.go:37)	MOVUPS	X0, "".a+328(SP)
	0x004b 00075 (sample/sample.go:37)	MOVQ	$0, "".~r1+248(SP)
	0x0057 00087 (sample/sample.go:37)	XORPS	X0, X0
	0x005a 00090 (sample/sample.go:37)	MOVUPS	X0, "".~r1+256(SP)
	0x0062 00098 (<unknown line number>)	NOP
	0x0062 00098 (sample/sample.go:15)	PCDATA	$0, $1
	0x0062 00098 (sample/sample.go:15)	MOVQ	"".a+320(SP), AX
	0x006a 00106 (sample/sample.go:15)	MOVQ	"".a+336(SP), CX
	0x0072 00114 (sample/sample.go:15)	PCDATA	$1, $0
	0x0072 00114 (sample/sample.go:15)	MOVQ	"".a+328(SP), DX
	0x007a 00122 (sample/sample.go:15)	MOVQ	AX, "".b+296(SP)
	0x0082 00130 (sample/sample.go:15)	MOVQ	DX, "".b+304(SP)
	0x008a 00138 (sample/sample.go:15)	MOVQ	CX, "".b+312(SP)
	0x0092 00146 (sample/sample.go:37)	PCDATA	$0, $0
	0x0092 00146 (sample/sample.go:37)	MOVQ	AX, "".~r1+248(SP)
	0x009a 00154 (sample/sample.go:37)	MOVQ	DX, "".~r1+256(SP)
	0x00a2 00162 (sample/sample.go:37)	MOVQ	CX, "".~r1+264(SP)
	0x00aa 00170 (sample/sample.go:37)	JMP	172
	0x00ac 00172 (sample/sample.go:38)	MOVW	$0, ""..autotmp_18+61(SP)
	0x00b3 00179 (sample/sample.go:38)	MOVB	$0, ""..autotmp_18+63(SP)
	0x00b8 00184 (sample/sample.go:38)	PCDATA	$0, $1
	0x00b8 00184 (sample/sample.go:38)	LEAQ	""..autotmp_18+61(SP), AX
	0x00bd 00189 (sample/sample.go:38)	PCDATA	$1, $2
	0x00bd 00189 (sample/sample.go:38)	MOVQ	AX, ""..autotmp_16+152(SP)
	0x00c5 00197 (sample/sample.go:38)	PCDATA	$0, $0
	0x00c5 00197 (sample/sample.go:38)	TESTB	AL, (AX)
	0x00c7 00199 (sample/sample.go:38)	MOVB	$97, ""..autotmp_18+61(SP)
	0x00cc 00204 (sample/sample.go:38)	PCDATA	$0, $1
	0x00cc 00204 (sample/sample.go:38)	MOVQ	""..autotmp_16+152(SP), AX
	0x00d4 00212 (sample/sample.go:38)	TESTB	AL, (AX)
	0x00d6 00214 (sample/sample.go:38)	PCDATA	$0, $0
	0x00d6 00214 (sample/sample.go:38)	MOVB	$98, 1(AX)
	0x00da 00218 (sample/sample.go:38)	PCDATA	$0, $1
	0x00da 00218 (sample/sample.go:38)	MOVQ	""..autotmp_16+152(SP), AX
	0x00e2 00226 (sample/sample.go:38)	TESTB	AL, (AX)
	0x00e4 00228 (sample/sample.go:38)	PCDATA	$0, $0
	0x00e4 00228 (sample/sample.go:38)	MOVB	$99, 2(AX)
	0x00e8 00232 (sample/sample.go:38)	PCDATA	$0, $1
	0x00e8 00232 (sample/sample.go:38)	PCDATA	$1, $0
	0x00e8 00232 (sample/sample.go:38)	MOVQ	""..autotmp_16+152(SP), AX
	0x00f0 00240 (sample/sample.go:38)	TESTB	AL, (AX)
	0x00f2 00242 (sample/sample.go:38)	JMP	244
	0x00f4 00244 (sample/sample.go:38)	MOVQ	AX, ""..autotmp_15+368(SP)
	0x00fc 00252 (sample/sample.go:38)	MOVQ	$3, ""..autotmp_15+376(SP)
	0x0108 00264 (sample/sample.go:38)	MOVQ	$3, ""..autotmp_15+384(SP)
	0x0114 00276 (sample/sample.go:38)	PCDATA	$0, $0
	0x0114 00276 (sample/sample.go:38)	PCDATA	$1, $3
	0x0114 00276 (sample/sample.go:38)	MOVQ	AX, "".a+344(SP)
	0x011c 00284 (sample/sample.go:38)	MOVQ	$3, "".a+352(SP)
	0x0128 00296 (sample/sample.go:38)	MOVQ	$3, "".a+360(SP)
	0x0134 00308 (sample/sample.go:38)	MOVQ	$0, "".~r1+224(SP)
	0x0140 00320 (sample/sample.go:38)	XORPS	X0, X0
	0x0143 00323 (sample/sample.go:38)	MOVUPS	X0, "".~r1+232(SP)
	0x014b 00331 (<unknown line number>)	NOP
	0x014b 00331 (sample/sample.go:15)	PCDATA	$0, $1
	0x014b 00331 (sample/sample.go:15)	MOVQ	"".a+344(SP), AX
	0x0153 00339 (sample/sample.go:15)	MOVQ	"".a+352(SP), CX
	0x015b 00347 (sample/sample.go:15)	PCDATA	$1, $0
	0x015b 00347 (sample/sample.go:15)	MOVQ	"".a+360(SP), DX
	0x0163 00355 (sample/sample.go:15)	MOVQ	AX, "".b+272(SP)
	0x016b 00363 (sample/sample.go:15)	MOVQ	CX, "".b+280(SP)
	0x0173 00371 (sample/sample.go:15)	MOVQ	DX, "".b+288(SP)
	0x017b 00379 (sample/sample.go:38)	PCDATA	$0, $0
	0x017b 00379 (sample/sample.go:38)	MOVQ	AX, "".~r1+224(SP)
	0x0183 00387 (sample/sample.go:38)	MOVQ	CX, "".~r1+232(SP)
	0x018b 00395 (sample/sample.go:38)	MOVQ	DX, "".~r1+240(SP)
	0x0193 00403 (sample/sample.go:38)	JMP	405
	0x0195 00405 (sample/sample.go:39)	MOVQ	$1, "".x+88(SP)
	0x019e 00414 (sample/sample.go:39)	JMP	416
	0x01a0 00416 (sample/sample.go:40)	MOVQ	$2, "".x+96(SP)
	0x01a9 00425 (sample/sample.go:40)	MOVQ	$0, "".~r1+72(SP)
	0x01b2 00434 (sample/sample.go:40)	MOVQ	"".x+96(SP), AX
	0x01b7 00439 (sample/sample.go:40)	MOVQ	AX, "".~r1+72(SP)
	0x01bc 00444 (sample/sample.go:40)	JMP	446
	0x01be 00446 (sample/sample.go:41)	MOVQ	$2, "".x+104(SP)
	0x01c7 00455 (sample/sample.go:41)	MOVQ	$3, "".y+80(SP)
	0x01d0 00464 (sample/sample.go:41)	MOVQ	$0, "".~r2+64(SP)
	0x01d9 00473 (<unknown line number>)	NOP
	0x01d9 00473 (sample/sample.go:28)	MOVQ	"".x+104(SP), AX
	0x01de 00478 (sample/sample.go:28)	ADDQ	"".y+80(SP), AX
	0x01e3 00483 (sample/sample.go:41)	MOVQ	AX, ""..autotmp_19+112(SP)
	0x01e8 00488 (sample/sample.go:41)	MOVQ	AX, "".~r2+64(SP)
	0x01ed 00493 (sample/sample.go:41)	JMP	495
	0x01ef 00495 (sample/sample.go:42)	PCDATA	$0, $1
	0x01ef 00495 (sample/sample.go:42)	PCDATA	$1, $4
	0x01ef 00495 (sample/sample.go:42)	LEAQ	go.string."hello"(SB), AX
	0x01f6 00502 (sample/sample.go:42)	PCDATA	$0, $0
	0x01f6 00502 (sample/sample.go:42)	MOVQ	AX, "".x+192(SP)
	0x01fe 00510 (sample/sample.go:42)	MOVQ	$5, "".x+200(SP)
	0x020a 00522 (sample/sample.go:42)	PCDATA	$0, $1
	0x020a 00522 (sample/sample.go:42)	PCDATA	$1, $5
	0x020a 00522 (sample/sample.go:42)	LEAQ	go.string." world"(SB), AX
	0x0211 00529 (sample/sample.go:42)	PCDATA	$0, $0
	0x0211 00529 (sample/sample.go:42)	MOVQ	AX, "".y+176(SP)
	0x0219 00537 (sample/sample.go:42)	MOVQ	$6, "".y+184(SP)
	0x0225 00549 (sample/sample.go:42)	XORPS	X0, X0
	0x0228 00552 (sample/sample.go:42)	MOVUPS	X0, "".~r2+160(SP)
	0x0230 00560 (<unknown line number>)	NOP
	0x0230 00560 (sample/sample.go:32)	PCDATA	$0, $1
	0x0230 00560 (sample/sample.go:32)	LEAQ	""..autotmp_21+120(SP), AX
	0x0235 00565 (sample/sample.go:32)	PCDATA	$0, $0
	0x0235 00565 (sample/sample.go:32)	MOVQ	AX, (SP)
	0x0239 00569 (sample/sample.go:32)	PCDATA	$0, $1
	0x0239 00569 (sample/sample.go:32)	MOVQ	"".x+192(SP), AX
	0x0241 00577 (sample/sample.go:32)	PCDATA	$1, $6
	0x0241 00577 (sample/sample.go:32)	MOVQ	"".x+200(SP), CX
	0x0249 00585 (sample/sample.go:32)	PCDATA	$0, $0
	0x0249 00585 (sample/sample.go:32)	MOVQ	AX, 8(SP)
	0x024e 00590 (sample/sample.go:32)	MOVQ	CX, 16(SP)
	0x0253 00595 (sample/sample.go:32)	PCDATA	$0, $1
	0x0253 00595 (sample/sample.go:32)	MOVQ	"".y+176(SP), AX
	0x025b 00603 (sample/sample.go:32)	PCDATA	$1, $0
	0x025b 00603 (sample/sample.go:32)	MOVQ	"".y+184(SP), CX
	0x0263 00611 (sample/sample.go:32)	PCDATA	$0, $0
	0x0263 00611 (sample/sample.go:32)	MOVQ	AX, 24(SP)
	0x0268 00616 (sample/sample.go:32)	MOVQ	CX, 32(SP)
	0x026d 00621 (sample/sample.go:32)	CALL	runtime.concatstring2(SB)
	0x0272 00626 (sample/sample.go:32)	PCDATA	$0, $1
	0x0272 00626 (sample/sample.go:32)	MOVQ	40(SP), AX
	0x0277 00631 (sample/sample.go:32)	MOVQ	48(SP), CX
	0x027c 00636 (sample/sample.go:42)	MOVQ	AX, ""..autotmp_20+208(SP)
	0x0284 00644 (sample/sample.go:42)	MOVQ	CX, ""..autotmp_20+216(SP)
	0x028c 00652 (sample/sample.go:42)	PCDATA	$0, $0
	0x028c 00652 (sample/sample.go:42)	MOVQ	AX, "".~r2+160(SP)
	0x0294 00660 (sample/sample.go:42)	MOVQ	CX, "".~r2+168(SP)
	0x029c 00668 (sample/sample.go:42)	JMP	670
	0x029e 00670 (sample/sample.go:42)	PCDATA	$0, $-1
	0x029e 00670 (sample/sample.go:42)	PCDATA	$1, $-1
	0x029e 00670 (sample/sample.go:42)	MOVQ	392(SP), BP
	0x02a6 00678 (sample/sample.go:42)	ADDQ	$400, SP
	0x02ad 00685 (sample/sample.go:42)	RET
	0x02ae 00686 (sample/sample.go:42)	NOP
	0x02ae 00686 (sample/sample.go:35)	PCDATA	$1, $-1
	0x02ae 00686 (sample/sample.go:35)	PCDATA	$0, $-2
	0x02ae 00686 (sample/sample.go:35)	CALL	runtime.morestack_noctxt(SB)
	0x02b3 00691 (sample/sample.go:35)	PCDATA	$0, $-1
	0x02b3 00691 (sample/sample.go:35)	JMP	0
	0x0000 64 48 8b 0c 25 00 00 00 00 48 8d 84 24 f0 fe ff  dH..%....H..$...
	0x0010 ff 48 3b 41 10 0f 86 93 02 00 00 48 81 ec 90 01  .H;A.......H....
	0x0020 00 00 48 89 ac 24 88 01 00 00 48 8d ac 24 88 01  ..H..$....H..$..
	0x0030 00 00 eb 00 48 c7 84 24 40 01 00 00 00 00 00 00  ....H..$@.......
	0x0040 0f 57 c0 0f 11 84 24 48 01 00 00 48 c7 84 24 f8  .W....$H...H..$.
	0x0050 00 00 00 00 00 00 00 0f 57 c0 0f 11 84 24 00 01  ........W....$..
	0x0060 00 00 48 8b 84 24 40 01 00 00 48 8b 8c 24 50 01  ..H..$@...H..$P.
	0x0070 00 00 48 8b 94 24 48 01 00 00 48 89 84 24 28 01  ..H..$H...H..$(.
	0x0080 00 00 48 89 94 24 30 01 00 00 48 89 8c 24 38 01  ..H..$0...H..$8.
	0x0090 00 00 48 89 84 24 f8 00 00 00 48 89 94 24 00 01  ..H..$....H..$..
	0x00a0 00 00 48 89 8c 24 08 01 00 00 eb 00 66 c7 44 24  ..H..$......f.D$
	0x00b0 3d 00 00 c6 44 24 3f 00 48 8d 44 24 3d 48 89 84  =...D$?.H.D$=H..
	0x00c0 24 98 00 00 00 84 00 c6 44 24 3d 61 48 8b 84 24  $.......D$=aH..$
	0x00d0 98 00 00 00 84 00 c6 40 01 62 48 8b 84 24 98 00  .......@.bH..$..
	0x00e0 00 00 84 00 c6 40 02 63 48 8b 84 24 98 00 00 00  .....@.cH..$....
	0x00f0 84 00 eb 00 48 89 84 24 70 01 00 00 48 c7 84 24  ....H..$p...H..$
	0x0100 78 01 00 00 03 00 00 00 48 c7 84 24 80 01 00 00  x.......H..$....
	0x0110 03 00 00 00 48 89 84 24 58 01 00 00 48 c7 84 24  ....H..$X...H..$
	0x0120 60 01 00 00 03 00 00 00 48 c7 84 24 68 01 00 00  `.......H..$h...
	0x0130 03 00 00 00 48 c7 84 24 e0 00 00 00 00 00 00 00  ....H..$........
	0x0140 0f 57 c0 0f 11 84 24 e8 00 00 00 48 8b 84 24 58  .W....$....H..$X
	0x0150 01 00 00 48 8b 8c 24 60 01 00 00 48 8b 94 24 68  ...H..$`...H..$h
	0x0160 01 00 00 48 89 84 24 10 01 00 00 48 89 8c 24 18  ...H..$....H..$.
	0x0170 01 00 00 48 89 94 24 20 01 00 00 48 89 84 24 e0  ...H..$ ...H..$.
	0x0180 00 00 00 48 89 8c 24 e8 00 00 00 48 89 94 24 f0  ...H..$....H..$.
	0x0190 00 00 00 eb 00 48 c7 44 24 58 01 00 00 00 eb 00  .....H.D$X......
	0x01a0 48 c7 44 24 60 02 00 00 00 48 c7 44 24 48 00 00  H.D$`....H.D$H..
	0x01b0 00 00 48 8b 44 24 60 48 89 44 24 48 eb 00 48 c7  ..H.D$`H.D$H..H.
	0x01c0 44 24 68 02 00 00 00 48 c7 44 24 50 03 00 00 00  D$h....H.D$P....
	0x01d0 48 c7 44 24 40 00 00 00 00 48 8b 44 24 68 48 03  H.D$@....H.D$hH.
	0x01e0 44 24 50 48 89 44 24 70 48 89 44 24 40 eb 00 48  D$PH.D$pH.D$@..H
	0x01f0 8d 05 00 00 00 00 48 89 84 24 c0 00 00 00 48 c7  ......H..$....H.
	0x0200 84 24 c8 00 00 00 05 00 00 00 48 8d 05 00 00 00  .$........H.....
	0x0210 00 48 89 84 24 b0 00 00 00 48 c7 84 24 b8 00 00  .H..$....H..$...
	0x0220 00 06 00 00 00 0f 57 c0 0f 11 84 24 a0 00 00 00  ......W....$....
	0x0230 48 8d 44 24 78 48 89 04 24 48 8b 84 24 c0 00 00  H.D$xH..$H..$...
	0x0240 00 48 8b 8c 24 c8 00 00 00 48 89 44 24 08 48 89  .H..$....H.D$.H.
	0x0250 4c 24 10 48 8b 84 24 b0 00 00 00 48 8b 8c 24 b8  L$.H..$....H..$.
	0x0260 00 00 00 48 89 44 24 18 48 89 4c 24 20 e8 00 00  ...H.D$.H.L$ ...
	0x0270 00 00 48 8b 44 24 28 48 8b 4c 24 30 48 89 84 24  ..H.D$(H.L$0H..$
	0x0280 d0 00 00 00 48 89 8c 24 d8 00 00 00 48 89 84 24  ....H..$....H..$
	0x0290 a0 00 00 00 48 89 8c 24 a8 00 00 00 eb 00 48 8b  ....H..$......H.
	0x02a0 ac 24 88 01 00 00 48 81 c4 90 01 00 00 c3 e8 00  .$....H.........
	0x02b0 00 00 00 e9 48 fd ff ff                          ....H...
	rel 5+4 t=17 TLS+0
	rel 498+4 t=16 go.string."hello"+0
	rel 525+4 t=16 go.string." world"+0
	rel 622+4 t=8 runtime.concatstring2+0
	rel 687+4 t=8 runtime.morestack_noctxt+0
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
	rel 27+4 t=30 gofile../mnt/sample/sample.go+0
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
go.loc."".sum SDWARFLOC size=0
go.info."".sum SDWARFINFO size=53
	0x0000 05 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0010 00 00 00 00 00 01 9c 12 00 00 00 00 01 9c 12 00  ................
	0x0020 00 00 00 02 91 08 0f 7e 72 32 00 01 1b 00 00 00  .......~r2......
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
	0x0000 04 02 03 15 14 6a 06 41 04 01 03 65 06 01        .....j.A...e..
go.loc."".concate SDWARFLOC size=0
go.info."".concate SDWARFINFO size=53
	0x0000 05 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0010 00 00 00 00 00 01 9c 12 00 00 00 00 01 9c 12 00  ................
	0x0020 00 00 00 02 91 10 0f 7e 72 32 00 01 1f 00 00 00  .......~r2......
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
	0x0000 04 02 03 19 14 0a a5 06 b9 06 42 06 5f 06 08 af  ..........B._...
	0x0010 06 41 06 08 4a 04 01 03 62 01                    .A..J...b.
go.string."hello" SRODATA dupok size=5
	0x0000 68 65 6c 6c 6f                                   hello
go.string." world" SRODATA dupok size=6
	0x0000 20 77 6f 72 6c 64                                 world
go.loc."".main SDWARFLOC size=0
go.info."".main SDWARFINFO size=212
	0x0000 03 22 22 2e 6d 61 69 6e 00 00 00 00 00 00 00 00  ."".main........
	0x0010 00 00 00 00 00 00 00 00 00 01 9c 00 00 00 00 01  ................
	0x0020 06 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0030 00 00 00 00 00 00 00 00 00 25 12 00 00 00 00 03  .........%......
	0x0040 91 a8 7f 0d 00 00 00 00 03 91 90 7f 00 06 00 00  ................
	0x0050 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0060 00 00 00 00 00 00 26 12 00 00 00 00 02 91 40 0d  ......&.......@.
	0x0070 00 00 00 00 03 91 f8 7e 00 06 00 00 00 00 00 00  .......~........
	0x0080 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0090 00 00 29 12 00 00 00 00 03 91 d0 7d 12 00 00 00  ..)........}....
	0x00a0 00 03 91 b8 7d 00 06 00 00 00 00 00 00 00 00 00  ....}...........
	0x00b0 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 2a  ...............*
	0x00c0 12 00 00 00 00 03 91 a8 7e 12 00 00 00 00 03 91  ........~.......
	0x00d0 98 7e 00 00                                      .~..
	rel 0+0 t=24 type.*[3]uint8+0
	rel 0+0 t=24 type.[32]uint8+0
	rel 0+0 t=24 type.[3]uint8+0
	rel 0+0 t=24 type.[]uint8+0
	rel 0+0 t=24 type.int+0
	rel 0+0 t=24 type.string+0
	rel 9+8 t=1 "".main+0
	rel 17+8 t=1 "".main+696
	rel 27+4 t=30 gofile../mnt/sample/sample.go+0
	rel 33+4 t=29 go.info."".slice$abstract+0
	rel 37+8 t=1 "".main+98
	rel 45+8 t=1 "".main+146
	rel 53+4 t=30 gofile../mnt/sample/sample.go+0
	rel 59+4 t=29 go.info."".slice$abstract+10
	rel 68+4 t=29 go.info."".slice$abstract+18
	rel 78+4 t=29 go.info."".slice$abstract+0
	rel 82+8 t=1 "".main+331
	rel 90+8 t=1 "".main+379
	rel 98+4 t=30 gofile../mnt/sample/sample.go+0
	rel 104+4 t=29 go.info."".slice$abstract+10
	rel 112+4 t=29 go.info."".slice$abstract+18
	rel 122+4 t=29 go.info."".sum$abstract+0
	rel 126+8 t=1 "".main+473
	rel 134+8 t=1 "".main+483
	rel 142+4 t=30 gofile../mnt/sample/sample.go+0
	rel 148+4 t=29 go.info."".sum$abstract+8
	rel 157+4 t=29 go.info."".sum$abstract+16
	rel 167+4 t=29 go.info."".concate$abstract+0
	rel 171+8 t=1 "".main+560
	rel 179+8 t=1 "".main+636
	rel 187+4 t=30 gofile../mnt/sample/sample.go+0
	rel 193+4 t=29 go.info."".concate$abstract+12
	rel 202+4 t=29 go.info."".concate$abstract+20
go.range."".main SDWARFRANGE size=0
go.debuglines."".main SDWARFMISC size=110
	0x0000 04 02 03 1d 14 0a 08 2d f6 24 06 87 06 08 03 6e  .......-.$.....n
	0x0010 6f 06 5f 06 08 03 11 b4 06 5f 06 c4 06 55 06 02  o._......_...U..
	0x0020 29 ff 06 5f 06 02 37 03 6d fb 06 5f 06 08 03 12  ).._..7.m.._....
	0x0030 b4 06 5f 06 c4 06 69 06 24 06 69 06 e2 06 69 06  .._...i.$.i...i.
	0x0040 03 77 bf 06 41 06 03 08 46 06 41 06 56 06 55 06  .w..A...F.A.V.U.
	0x0050 02 22 03 7a fb 06 41 06 02 20 ff 06 41 06 03 05  .".z..A.. ..A...
	0x0060 78 06 5f 06 08 23 03 7d ab 04 01 03 5e 01        x._..#.}....^.
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
gclocals·ad9a9b86035be5675884f71ffff3e114 SRODATA dupok size=36
	0x0000 07 00 00 00 1e 00 00 00 00 00 00 00 00 00 20 00  .............. .
	0x0010 01 00 00 00 00 00 00 01 20 00 00 00 28 00 00 00  ........ ...(...
	0x0020 08 00 00 00                                      ....
