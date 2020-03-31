"".min STEXT nosplit size=1 args=0x0 locals=0x0
	0x0000 00000 (sample/sample.go:5)	TEXT	"".min(SB), NOSPLIT|ABIInternal, $0-0
	0x0000 00000 (sample/sample.go:5)	PCDATA	$0, $-2
	0x0000 00000 (sample/sample.go:5)	PCDATA	$1, $-2
	0x0000 00000 (sample/sample.go:5)	FUNCDATA	$0, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0000 00000 (sample/sample.go:5)	FUNCDATA	$1, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0000 00000 (sample/sample.go:5)	FUNCDATA	$2, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0000 00000 (sample/sample.go:7)	PCDATA	$0, $-1
	0x0000 00000 (sample/sample.go:7)	PCDATA	$1, $-1
	0x0000 00000 (sample/sample.go:7)	RET
	0x0000 c3                                               .
"".arg1 STEXT nosplit size=1 args=0x8 locals=0x0
	0x0000 00000 (sample/sample.go:9)	TEXT	"".arg1(SB), NOSPLIT|ABIInternal, $0-8
	0x0000 00000 (sample/sample.go:9)	PCDATA	$0, $-2
	0x0000 00000 (sample/sample.go:9)	PCDATA	$1, $-2
	0x0000 00000 (sample/sample.go:9)	FUNCDATA	$0, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0000 00000 (sample/sample.go:9)	FUNCDATA	$1, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0000 00000 (sample/sample.go:9)	FUNCDATA	$2, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0000 00000 (sample/sample.go:11)	PCDATA	$0, $-1
	0x0000 00000 (sample/sample.go:11)	PCDATA	$1, $-1
	0x0000 00000 (sample/sample.go:11)	RET
	0x0000 c3                                               .
"".arg1ret1 STEXT nosplit size=20 args=0x10 locals=0x0
	0x0000 00000 (sample/sample.go:13)	TEXT	"".arg1ret1(SB), NOSPLIT|ABIInternal, $0-16
	0x0000 00000 (sample/sample.go:13)	PCDATA	$0, $-2
	0x0000 00000 (sample/sample.go:13)	PCDATA	$1, $-2
	0x0000 00000 (sample/sample.go:13)	FUNCDATA	$0, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0000 00000 (sample/sample.go:13)	FUNCDATA	$1, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0000 00000 (sample/sample.go:13)	FUNCDATA	$2, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0000 00000 (sample/sample.go:13)	PCDATA	$0, $0
	0x0000 00000 (sample/sample.go:13)	PCDATA	$1, $0
	0x0000 00000 (sample/sample.go:13)	MOVQ	$0, "".~r1+16(SP)
	0x0009 00009 (sample/sample.go:14)	MOVQ	"".x+8(SP), AX
	0x000e 00014 (sample/sample.go:14)	MOVQ	AX, "".~r1+16(SP)
	0x0013 00019 (sample/sample.go:14)	RET
	0x0000 48 c7 44 24 10 00 00 00 00 48 8b 44 24 08 48 89  H.D$.....H.D$.H.
	0x0010 44 24 10 c3                                      D$..
"".sum STEXT nosplit size=25 args=0x18 locals=0x0
	0x0000 00000 (sample/sample.go:17)	TEXT	"".sum(SB), NOSPLIT|ABIInternal, $0-24
	0x0000 00000 (sample/sample.go:17)	PCDATA	$0, $-2
	0x0000 00000 (sample/sample.go:17)	PCDATA	$1, $-2
	0x0000 00000 (sample/sample.go:17)	FUNCDATA	$0, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0000 00000 (sample/sample.go:17)	FUNCDATA	$1, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0000 00000 (sample/sample.go:17)	FUNCDATA	$2, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0000 00000 (sample/sample.go:17)	PCDATA	$0, $0
	0x0000 00000 (sample/sample.go:17)	PCDATA	$1, $0
	0x0000 00000 (sample/sample.go:17)	MOVQ	$0, "".~r2+24(SP)
	0x0009 00009 (sample/sample.go:18)	MOVQ	"".x+8(SP), AX
	0x000e 00014 (sample/sample.go:18)	ADDQ	"".y+16(SP), AX
	0x0013 00019 (sample/sample.go:18)	MOVQ	AX, "".~r2+24(SP)
	0x0018 00024 (sample/sample.go:18)	RET
	0x0000 48 c7 44 24 18 00 00 00 00 48 8b 44 24 08 48 03  H.D$.....H.D$.H.
	0x0010 44 24 10 48 89 44 24 18 c3                       D$.H.D$..
"".concate STEXT size=127 args=0x30 locals=0x40
	0x0000 00000 (sample/sample.go:21)	TEXT	"".concate(SB), ABIInternal, $64-48
	0x0000 00000 (sample/sample.go:21)	MOVQ	(TLS), CX
	0x0009 00009 (sample/sample.go:21)	CMPQ	SP, 16(CX)
	0x000d 00013 (sample/sample.go:21)	PCDATA	$0, $-2
	0x000d 00013 (sample/sample.go:21)	JLS	120
	0x000f 00015 (sample/sample.go:21)	PCDATA	$0, $-1
	0x000f 00015 (sample/sample.go:21)	SUBQ	$64, SP
	0x0013 00019 (sample/sample.go:21)	MOVQ	BP, 56(SP)
	0x0018 00024 (sample/sample.go:21)	LEAQ	56(SP), BP
	0x001d 00029 (sample/sample.go:21)	PCDATA	$0, $-2
	0x001d 00029 (sample/sample.go:21)	PCDATA	$1, $-2
	0x001d 00029 (sample/sample.go:21)	FUNCDATA	$0, gclocals·2625d1fdbbaf79a2e52296235cb6527c(SB)
	0x001d 00029 (sample/sample.go:21)	FUNCDATA	$1, gclocals·f6bd6b3389b872033d462029172c8612(SB)
	0x001d 00029 (sample/sample.go:21)	FUNCDATA	$2, gclocals·1cf923758aae2e428391d1783fe59973(SB)
	0x001d 00029 (sample/sample.go:21)	PCDATA	$0, $0
	0x001d 00029 (sample/sample.go:21)	PCDATA	$1, $0
	0x001d 00029 (sample/sample.go:21)	XORPS	X0, X0
	0x0020 00032 (sample/sample.go:21)	MOVUPS	X0, "".~r2+104(SP)
	0x0025 00037 (sample/sample.go:22)	MOVQ	$0, (SP)
	0x002d 00045 (sample/sample.go:22)	PCDATA	$0, $1
	0x002d 00045 (sample/sample.go:22)	MOVQ	"".x+72(SP), AX
	0x0032 00050 (sample/sample.go:22)	PCDATA	$1, $1
	0x0032 00050 (sample/sample.go:22)	MOVQ	"".x+80(SP), CX
	0x0037 00055 (sample/sample.go:22)	PCDATA	$0, $0
	0x0037 00055 (sample/sample.go:22)	MOVQ	AX, 8(SP)
	0x003c 00060 (sample/sample.go:22)	MOVQ	CX, 16(SP)
	0x0041 00065 (sample/sample.go:22)	PCDATA	$0, $1
	0x0041 00065 (sample/sample.go:22)	MOVQ	"".y+88(SP), AX
	0x0046 00070 (sample/sample.go:22)	PCDATA	$1, $2
	0x0046 00070 (sample/sample.go:22)	MOVQ	"".y+96(SP), CX
	0x004b 00075 (sample/sample.go:22)	PCDATA	$0, $0
	0x004b 00075 (sample/sample.go:22)	MOVQ	AX, 24(SP)
	0x0050 00080 (sample/sample.go:22)	MOVQ	CX, 32(SP)
	0x0055 00085 (sample/sample.go:22)	CALL	runtime.concatstring2(SB)
	0x005a 00090 (sample/sample.go:22)	MOVQ	48(SP), AX
	0x005f 00095 (sample/sample.go:22)	PCDATA	$0, $2
	0x005f 00095 (sample/sample.go:22)	MOVQ	40(SP), CX
	0x0064 00100 (sample/sample.go:22)	PCDATA	$0, $0
	0x0064 00100 (sample/sample.go:22)	PCDATA	$1, $3
	0x0064 00100 (sample/sample.go:22)	MOVQ	CX, "".~r2+104(SP)
	0x0069 00105 (sample/sample.go:22)	MOVQ	AX, "".~r2+112(SP)
	0x006e 00110 (sample/sample.go:22)	MOVQ	56(SP), BP
	0x0073 00115 (sample/sample.go:22)	ADDQ	$64, SP
	0x0077 00119 (sample/sample.go:22)	RET
	0x0078 00120 (sample/sample.go:22)	NOP
	0x0078 00120 (sample/sample.go:21)	PCDATA	$1, $-1
	0x0078 00120 (sample/sample.go:21)	PCDATA	$0, $-2
	0x0078 00120 (sample/sample.go:21)	CALL	runtime.morestack_noctxt(SB)
	0x007d 00125 (sample/sample.go:21)	PCDATA	$0, $-1
	0x007d 00125 (sample/sample.go:21)	JMP	0
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
"".main STEXT size=340 args=0x0 locals=0xd8
	0x0000 00000 (sample/sample.go:25)	TEXT	"".main(SB), ABIInternal, $216-0
	0x0000 00000 (sample/sample.go:25)	MOVQ	(TLS), CX
	0x0009 00009 (sample/sample.go:25)	LEAQ	-88(SP), AX
	0x000e 00014 (sample/sample.go:25)	CMPQ	AX, 16(CX)
	0x0012 00018 (sample/sample.go:25)	PCDATA	$0, $-2
	0x0012 00018 (sample/sample.go:25)	JLS	330
	0x0018 00024 (sample/sample.go:25)	PCDATA	$0, $-1
	0x0018 00024 (sample/sample.go:25)	SUBQ	$216, SP
	0x001f 00031 (sample/sample.go:25)	MOVQ	BP, 208(SP)
	0x0027 00039 (sample/sample.go:25)	LEAQ	208(SP), BP
	0x002f 00047 (sample/sample.go:25)	PCDATA	$0, $-2
	0x002f 00047 (sample/sample.go:25)	PCDATA	$1, $-2
	0x002f 00047 (sample/sample.go:25)	FUNCDATA	$0, gclocals·f6bd6b3389b872033d462029172c8612(SB)
	0x002f 00047 (sample/sample.go:25)	FUNCDATA	$1, gclocals·c1a2e878a9a3dc38b4d01d6a1bc52e0a(SB)
	0x002f 00047 (sample/sample.go:25)	FUNCDATA	$2, gclocals·1cf923758aae2e428391d1783fe59973(SB)
	0x002f 00047 (sample/sample.go:26)	PCDATA	$0, $-1
	0x002f 00047 (sample/sample.go:26)	PCDATA	$1, $-1
	0x002f 00047 (sample/sample.go:26)	JMP	49
	0x0031 00049 (sample/sample.go:27)	PCDATA	$0, $0
	0x0031 00049 (sample/sample.go:27)	PCDATA	$1, $0
	0x0031 00049 (sample/sample.go:27)	MOVQ	$1, "".x+96(SP)
	0x003a 00058 (sample/sample.go:27)	JMP	60
	0x003c 00060 (sample/sample.go:28)	MOVQ	$2, "".x+88(SP)
	0x0045 00069 (sample/sample.go:28)	MOVQ	$0, "".~r1+64(SP)
	0x004e 00078 (sample/sample.go:28)	MOVQ	"".x+88(SP), AX
	0x0053 00083 (sample/sample.go:28)	MOVQ	AX, "".~r1+64(SP)
	0x0058 00088 (sample/sample.go:28)	JMP	90
	0x005a 00090 (sample/sample.go:29)	MOVQ	$2, "".x+80(SP)
	0x0063 00099 (sample/sample.go:29)	MOVQ	$3, "".y+72(SP)
	0x006c 00108 (sample/sample.go:29)	MOVQ	$0, "".~r2+56(SP)
	0x0075 00117 (<unknown line number>)	NOP
	0x0075 00117 (sample/sample.go:18)	MOVQ	"".x+80(SP), AX
	0x007a 00122 (sample/sample.go:18)	ADDQ	"".y+72(SP), AX
	0x007f 00127 (sample/sample.go:29)	MOVQ	AX, ""..autotmp_9+104(SP)
	0x0084 00132 (sample/sample.go:29)	MOVQ	AX, "".~r2+56(SP)
	0x0089 00137 (sample/sample.go:29)	JMP	139
	0x008b 00139 (sample/sample.go:30)	PCDATA	$0, $1
	0x008b 00139 (sample/sample.go:30)	PCDATA	$1, $1
	0x008b 00139 (sample/sample.go:30)	LEAQ	go.string."hello"(SB), AX
	0x0092 00146 (sample/sample.go:30)	PCDATA	$0, $0
	0x0092 00146 (sample/sample.go:30)	MOVQ	AX, "".x+176(SP)
	0x009a 00154 (sample/sample.go:30)	MOVQ	$5, "".x+184(SP)
	0x00a6 00166 (sample/sample.go:30)	PCDATA	$0, $1
	0x00a6 00166 (sample/sample.go:30)	PCDATA	$1, $2
	0x00a6 00166 (sample/sample.go:30)	LEAQ	go.string." world"(SB), AX
	0x00ad 00173 (sample/sample.go:30)	PCDATA	$0, $0
	0x00ad 00173 (sample/sample.go:30)	MOVQ	AX, "".y+160(SP)
	0x00b5 00181 (sample/sample.go:30)	MOVQ	$6, "".y+168(SP)
	0x00c1 00193 (sample/sample.go:30)	XORPS	X0, X0
	0x00c4 00196 (sample/sample.go:30)	MOVUPS	X0, "".~r2+144(SP)
	0x00cc 00204 (<unknown line number>)	NOP
	0x00cc 00204 (sample/sample.go:22)	PCDATA	$0, $1
	0x00cc 00204 (sample/sample.go:22)	LEAQ	""..autotmp_11+112(SP), AX
	0x00d1 00209 (sample/sample.go:22)	PCDATA	$0, $0
	0x00d1 00209 (sample/sample.go:22)	MOVQ	AX, (SP)
	0x00d5 00213 (sample/sample.go:22)	MOVQ	"".x+184(SP), AX
	0x00dd 00221 (sample/sample.go:22)	PCDATA	$0, $2
	0x00dd 00221 (sample/sample.go:22)	PCDATA	$1, $3
	0x00dd 00221 (sample/sample.go:22)	MOVQ	"".x+176(SP), CX
	0x00e5 00229 (sample/sample.go:22)	PCDATA	$0, $0
	0x00e5 00229 (sample/sample.go:22)	MOVQ	CX, 8(SP)
	0x00ea 00234 (sample/sample.go:22)	MOVQ	AX, 16(SP)
	0x00ef 00239 (sample/sample.go:22)	PCDATA	$0, $1
	0x00ef 00239 (sample/sample.go:22)	MOVQ	"".y+160(SP), AX
	0x00f7 00247 (sample/sample.go:22)	PCDATA	$1, $0
	0x00f7 00247 (sample/sample.go:22)	MOVQ	"".y+168(SP), CX
	0x00ff 00255 (sample/sample.go:22)	PCDATA	$0, $0
	0x00ff 00255 (sample/sample.go:22)	MOVQ	AX, 24(SP)
	0x0104 00260 (sample/sample.go:22)	MOVQ	CX, 32(SP)
	0x0109 00265 (sample/sample.go:22)	CALL	runtime.concatstring2(SB)
	0x010e 00270 (sample/sample.go:22)	PCDATA	$0, $1
	0x010e 00270 (sample/sample.go:22)	MOVQ	40(SP), AX
	0x0113 00275 (sample/sample.go:22)	MOVQ	48(SP), CX
	0x0118 00280 (sample/sample.go:30)	MOVQ	AX, ""..autotmp_10+192(SP)
	0x0120 00288 (sample/sample.go:30)	MOVQ	CX, ""..autotmp_10+200(SP)
	0x0128 00296 (sample/sample.go:30)	PCDATA	$0, $0
	0x0128 00296 (sample/sample.go:30)	MOVQ	AX, "".~r2+144(SP)
	0x0130 00304 (sample/sample.go:30)	MOVQ	CX, "".~r2+152(SP)
	0x0138 00312 (sample/sample.go:30)	JMP	314
	0x013a 00314 (sample/sample.go:30)	PCDATA	$0, $-1
	0x013a 00314 (sample/sample.go:30)	PCDATA	$1, $-1
	0x013a 00314 (sample/sample.go:30)	MOVQ	208(SP), BP
	0x0142 00322 (sample/sample.go:30)	ADDQ	$216, SP
	0x0149 00329 (sample/sample.go:30)	RET
	0x014a 00330 (sample/sample.go:30)	NOP
	0x014a 00330 (sample/sample.go:25)	PCDATA	$1, $-1
	0x014a 00330 (sample/sample.go:25)	PCDATA	$0, $-2
	0x014a 00330 (sample/sample.go:25)	CALL	runtime.morestack_noctxt(SB)
	0x014f 00335 (sample/sample.go:25)	PCDATA	$0, $-1
	0x014f 00335 (sample/sample.go:25)	JMP	0
	0x0000 64 48 8b 0c 25 00 00 00 00 48 8d 44 24 a8 48 3b  dH..%....H.D$.H;
	0x0010 41 10 0f 86 32 01 00 00 48 81 ec d8 00 00 00 48  A...2...H......H
	0x0020 89 ac 24 d0 00 00 00 48 8d ac 24 d0 00 00 00 eb  ..$....H..$.....
	0x0030 00 48 c7 44 24 60 01 00 00 00 eb 00 48 c7 44 24  .H.D$`......H.D$
	0x0040 58 02 00 00 00 48 c7 44 24 40 00 00 00 00 48 8b  X....H.D$@....H.
	0x0050 44 24 58 48 89 44 24 40 eb 00 48 c7 44 24 50 02  D$XH.D$@..H.D$P.
	0x0060 00 00 00 48 c7 44 24 48 03 00 00 00 48 c7 44 24  ...H.D$H....H.D$
	0x0070 38 00 00 00 00 48 8b 44 24 50 48 03 44 24 48 48  8....H.D$PH.D$HH
	0x0080 89 44 24 68 48 89 44 24 38 eb 00 48 8d 05 00 00  .D$hH.D$8..H....
	0x0090 00 00 48 89 84 24 b0 00 00 00 48 c7 84 24 b8 00  ..H..$....H..$..
	0x00a0 00 00 05 00 00 00 48 8d 05 00 00 00 00 48 89 84  ......H......H..
	0x00b0 24 a0 00 00 00 48 c7 84 24 a8 00 00 00 06 00 00  $....H..$.......
	0x00c0 00 0f 57 c0 0f 11 84 24 90 00 00 00 48 8d 44 24  ..W....$....H.D$
	0x00d0 70 48 89 04 24 48 8b 84 24 b8 00 00 00 48 8b 8c  pH..$H..$....H..
	0x00e0 24 b0 00 00 00 48 89 4c 24 08 48 89 44 24 10 48  $....H.L$.H.D$.H
	0x00f0 8b 84 24 a0 00 00 00 48 8b 8c 24 a8 00 00 00 48  ..$....H..$....H
	0x0100 89 44 24 18 48 89 4c 24 20 e8 00 00 00 00 48 8b  .D$.H.L$ .....H.
	0x0110 44 24 28 48 8b 4c 24 30 48 89 84 24 c0 00 00 00  D$(H.L$0H..$....
	0x0120 48 89 8c 24 c8 00 00 00 48 89 84 24 90 00 00 00  H..$....H..$....
	0x0130 48 89 8c 24 98 00 00 00 eb 00 48 8b ac 24 d0 00  H..$......H..$..
	0x0140 00 00 48 81 c4 d8 00 00 00 c3 e8 00 00 00 00 e9  ..H.............
	0x0150 ac fe ff ff                                      ....
	rel 5+4 t=17 TLS+0
	rel 142+4 t=16 go.string."hello"+0
	rel 169+4 t=16 go.string." world"+0
	rel 266+4 t=8 runtime.concatstring2+0
	rel 331+4 t=8 runtime.morestack_noctxt+0
go.cuinfo.packagename. SDWARFINFO dupok size=0
	0x0000 6d 61 69 6e                                      main
go.info."".min$abstract SDWARFINFO dupok size=9
	0x0000 04 2e 6d 69 6e 00 01 01 00                       ..min....
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
go.debuglines."".min SDWARFMISC size=10
	0x0000 04 02 03 01 14 04 01 03 7a 01                    ........z.
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
	0x0000 04 02 03 05 14 04 01 03 76 01                    ........v.
go.loc."".arg1ret1 SDWARFLOC size=0
go.info."".arg1ret1 SDWARFINFO size=45
	0x0000 05 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0010 00 00 00 00 00 01 9c 12 00 00 00 00 01 9c 0f 7e  ...............~
	0x0020 72 31 00 01 0d 00 00 00 00 02 91 08 00           r1...........
	rel 0+0 t=24 type.int+0
	rel 1+4 t=29 go.info."".arg1ret1$abstract+0
	rel 5+8 t=1 "".arg1ret1+0
	rel 13+8 t=1 "".arg1ret1+20
	rel 24+4 t=29 go.info."".arg1ret1$abstract+13
	rel 37+4 t=29 go.info.int+0
go.range."".arg1ret1 SDWARFRANGE size=0
go.debuglines."".arg1ret1 SDWARFMISC size=14
	0x0000 04 02 03 07 14 6a 06 41 04 01 03 73 06 01        .....j.A...s..
go.loc."".sum SDWARFLOC size=0
go.info."".sum SDWARFINFO size=53
	0x0000 05 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0010 00 00 00 00 00 01 9c 12 00 00 00 00 01 9c 12 00  ................
	0x0020 00 00 00 02 91 08 0f 7e 72 32 00 01 11 00 00 00  .......~r2......
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
	0x0000 04 02 03 0b 14 6a 06 41 04 01 03 6f 06 01        .....j.A...o..
go.loc."".concate SDWARFLOC size=0
go.info."".concate SDWARFINFO size=53
	0x0000 05 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0010 00 00 00 00 00 01 9c 12 00 00 00 00 01 9c 12 00  ................
	0x0020 00 00 00 02 91 10 0f 7e 72 32 00 01 15 00 00 00  .......~r2......
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
	0x0000 04 02 03 0f 14 0a a5 06 b9 06 42 06 5f 06 08 af  ..........B._...
	0x0010 06 41 06 08 4a 04 01 03 6c 01                    .A..J...l.
go.string."hello" SRODATA dupok size=5
	0x0000 68 65 6c 6c 6f                                   hello
go.string." world" SRODATA dupok size=6
	0x0000 20 77 6f 72 6c 64                                 world
go.loc."".main SDWARFLOC size=0
go.info."".main SDWARFINFO size=121
	0x0000 03 22 22 2e 6d 61 69 6e 00 00 00 00 00 00 00 00  ."".main........
	0x0010 00 00 00 00 00 00 00 00 00 01 9c 00 00 00 00 01  ................
	0x0020 06 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0030 00 00 00 00 00 00 00 00 00 1d 12 00 00 00 00 03  ................
	0x0040 91 f0 7e 12 00 00 00 00 03 91 e8 7e 00 06 00 00  ..~........~....
	0x0050 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0060 00 00 00 00 00 00 1e 12 00 00 00 00 02 91 50 12  ..............P.
	0x0070 00 00 00 00 02 91 40 00 00                       ......@..
	rel 0+0 t=24 type.[32]uint8+0
	rel 0+0 t=24 type.int+0
	rel 0+0 t=24 type.string+0
	rel 9+8 t=1 "".main+0
	rel 17+8 t=1 "".main+340
	rel 27+4 t=30 gofile../mnt/sample/sample.go+0
	rel 33+4 t=29 go.info."".sum$abstract+0
	rel 37+8 t=1 "".main+117
	rel 45+8 t=1 "".main+127
	rel 53+4 t=30 gofile../mnt/sample/sample.go+0
	rel 59+4 t=29 go.info."".sum$abstract+8
	rel 68+4 t=29 go.info."".sum$abstract+16
	rel 78+4 t=29 go.info."".concate$abstract+0
	rel 82+8 t=1 "".main+204
	rel 90+8 t=1 "".main+280
	rel 98+4 t=30 gofile../mnt/sample/sample.go+0
	rel 104+4 t=29 go.info."".concate$abstract+12
	rel 112+4 t=29 go.info."".concate$abstract+20
go.range."".main SDWARFRANGE size=0
go.debuglines."".main SDWARFMISC size=66
	0x0000 04 02 03 13 14 0a ff f6 24 06 69 06 24 06 69 06  ........$.i.$.i.
	0x0010 e2 06 69 06 03 79 bf 06 41 06 03 06 46 06 41 06  ..i..y..A...F.A.
	0x0020 56 06 55 06 02 22 03 7c fb 06 41 06 02 20 ff 06  V.U..".|..A.. ..
	0x0030 41 06 03 03 78 06 5f 06 08 23 03 7f ab 04 01 03  A...x._..#......
	0x0040 68 01                                            h.
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
runtime.gcbits. SRODATA dupok size=0
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
gclocals·2625d1fdbbaf79a2e52296235cb6527c SRODATA dupok size=12
	0x0000 04 00 00 00 05 00 00 00 05 04 00 10              ............
gclocals·f6bd6b3389b872033d462029172c8612 SRODATA dupok size=8
	0x0000 04 00 00 00 00 00 00 00                          ........
gclocals·1cf923758aae2e428391d1783fe59973 SRODATA dupok size=11
	0x0000 03 00 00 00 02 00 00 00 00 01 02                 ...........
gclocals·c1a2e878a9a3dc38b4d01d6a1bc52e0a SRODATA dupok size=12
	0x0000 04 00 00 00 08 00 00 00 00 10 14 04              ............
