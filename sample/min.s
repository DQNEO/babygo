"".main STEXT size=86 args=0x0 locals=0x18 funcid=0x0
	0x0000 00000 (min.go:4)	TEXT	"".main(SB), ABIInternal, $24-0
	0x0000 00000 (min.go:4)	MOVQ	(TLS), CX
	0x0009 00009 (min.go:4)	CMPQ	SP, 16(CX)
	0x000d 00013 (min.go:4)	PCDATA	$0, $-2
	0x000d 00013 (min.go:4)	JLS	79
	0x000f 00015 (min.go:4)	PCDATA	$0, $-1
	0x000f 00015 (min.go:4)	SUBQ	$24, SP
	0x0013 00019 (min.go:4)	MOVQ	BP, 16(SP)
	0x0018 00024 (min.go:4)	LEAQ	16(SP), BP
	0x001d 00029 (min.go:4)	FUNCDATA	$0, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x001d 00029 (min.go:4)	FUNCDATA	$1, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x001d 00029 (min.go:5)	PCDATA	$1, $0
	0x001d 00029 (min.go:5)	NOP
	0x0020 00032 (min.go:5)	CALL	runtime.printlock(SB)
	0x0025 00037 (min.go:5)	LEAQ	go.string."hello world\n"(SB), AX
	0x002c 00044 (min.go:5)	MOVQ	AX, (SP)
	0x0030 00048 (min.go:5)	MOVQ	$12, 8(SP)
	0x0039 00057 (min.go:5)	CALL	runtime.printstring(SB)
	0x003e 00062 (min.go:5)	NOP
	0x0040 00064 (min.go:5)	CALL	runtime.printunlock(SB)
	0x0045 00069 (min.go:6)	MOVQ	16(SP), BP
	0x004a 00074 (min.go:6)	ADDQ	$24, SP
	0x004e 00078 (min.go:6)	RET
	0x004f 00079 (min.go:6)	NOP
	0x004f 00079 (min.go:4)	PCDATA	$1, $-1
	0x004f 00079 (min.go:4)	PCDATA	$0, $-2
	0x004f 00079 (min.go:4)	CALL	runtime.morestack_noctxt(SB)
	0x0054 00084 (min.go:4)	PCDATA	$0, $-1
	0x0054 00084 (min.go:4)	JMP	0
	0x0000 65 48 8b 0c 25 00 00 00 00 48 3b 61 10 76 40 48  eH..%....H;a.v@H
	0x0010 83 ec 18 48 89 6c 24 10 48 8d 6c 24 10 0f 1f 00  ...H.l$.H.l$....
	0x0020 e8 00 00 00 00 48 8d 05 00 00 00 00 48 89 04 24  .....H......H..$
	0x0030 48 c7 44 24 08 0c 00 00 00 e8 00 00 00 00 66 90  H.D$..........f.
	0x0040 e8 00 00 00 00 48 8b 6c 24 10 48 83 c4 18 c3 e8  .....H.l$.H.....
	0x0050 00 00 00 00 eb aa                                ......
	rel 5+4 t=17 TLS+0
	rel 33+4 t=8 runtime.printlock+0
	rel 40+4 t=16 go.string."hello world\n"+0
	rel 58+4 t=8 runtime.printstring+0
	rel 65+4 t=8 runtime.printunlock+0
	rel 80+4 t=8 runtime.morestack_noctxt+0
go.cuinfo.packagename. SDWARFCUINFO dupok size=0
	0x0000 6d 61 69 6e                                      main
""..inittask SNOPTRDATA size=24
	0x0000 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0010 00 00 00 00 00 00 00 00                          ........
go.string."hello world" SRODATA dupok size=11
	0x0000 68 65 6c 6c 6f 20 77 6f 72 6c 64                 hello world
go.string."hello world\n" SRODATA dupok size=12
	0x0000 68 65 6c 6c 6f 20 77 6f 72 6c 64 0a              hello world.
gclocals·33cdeccccebe80329f1fdbee7f5874cb SRODATA dupok size=8
	0x0000 01 00 00 00 00 00 00 00                          ........
