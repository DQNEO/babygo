"".main STEXT size=81 args=0x0 locals=0x18
	0x0000 00000 (sample/min.go:4)	TEXT	"".main(SB), ABIInternal, $24-0
	0x0000 00000 (sample/min.go:4)	MOVQ	(TLS), CX
	0x0009 00009 (sample/min.go:4)	CMPQ	SP, 16(CX)
	0x000d 00013 (sample/min.go:4)	PCDATA	$0, $-2
	0x000d 00013 (sample/min.go:4)	JLS	74
	0x000f 00015 (sample/min.go:4)	PCDATA	$0, $-1
	0x000f 00015 (sample/min.go:4)	SUBQ	$24, SP
	0x0013 00019 (sample/min.go:4)	MOVQ	BP, 16(SP)
	0x0018 00024 (sample/min.go:4)	LEAQ	16(SP), BP
	0x001d 00029 (sample/min.go:4)	PCDATA	$0, $-2
	0x001d 00029 (sample/min.go:4)	PCDATA	$1, $-2
	0x001d 00029 (sample/min.go:4)	FUNCDATA	$0, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x001d 00029 (sample/min.go:4)	FUNCDATA	$1, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x001d 00029 (sample/min.go:4)	FUNCDATA	$2, gclocals·9fb7f0986f647f17cb53dda1484e0f7a(SB)
	0x001d 00029 (sample/min.go:5)	PCDATA	$0, $0
	0x001d 00029 (sample/min.go:5)	PCDATA	$1, $0
	0x001d 00029 (sample/min.go:5)	CALL	runtime.printlock(SB)
	0x0022 00034 (sample/min.go:5)	PCDATA	$0, $1
	0x0022 00034 (sample/min.go:5)	LEAQ	go.string."hello world\n"(SB), AX
	0x0029 00041 (sample/min.go:5)	PCDATA	$0, $0
	0x0029 00041 (sample/min.go:5)	MOVQ	AX, (SP)
	0x002d 00045 (sample/min.go:5)	MOVQ	$12, 8(SP)
	0x0036 00054 (sample/min.go:5)	CALL	runtime.printstring(SB)
	0x003b 00059 (sample/min.go:5)	CALL	runtime.printunlock(SB)
	0x0040 00064 (sample/min.go:6)	MOVQ	16(SP), BP
	0x0045 00069 (sample/min.go:6)	ADDQ	$24, SP
	0x0049 00073 (sample/min.go:6)	RET
	0x004a 00074 (sample/min.go:6)	NOP
	0x004a 00074 (sample/min.go:4)	PCDATA	$1, $-1
	0x004a 00074 (sample/min.go:4)	PCDATA	$0, $-2
	0x004a 00074 (sample/min.go:4)	CALL	runtime.morestack_noctxt(SB)
	0x004f 00079 (sample/min.go:4)	PCDATA	$0, $-1
	0x004f 00079 (sample/min.go:4)	JMP	0
	0x0000 64 48 8b 0c 25 00 00 00 00 48 3b 61 10 76 3b 48  dH..%....H;a.v;H
	0x0010 83 ec 18 48 89 6c 24 10 48 8d 6c 24 10 e8 00 00  ...H.l$.H.l$....
	0x0020 00 00 48 8d 05 00 00 00 00 48 89 04 24 48 c7 44  ..H......H..$H.D
	0x0030 24 08 0c 00 00 00 e8 00 00 00 00 e8 00 00 00 00  $...............
	0x0040 48 8b 6c 24 10 48 83 c4 18 c3 e8 00 00 00 00 eb  H.l$.H..........
	0x0050 af                                               .
	rel 5+4 t=17 TLS+0
	rel 30+4 t=8 runtime.printlock+0
	rel 37+4 t=16 go.string."hello world\n"+0
	rel 55+4 t=8 runtime.printstring+0
	rel 60+4 t=8 runtime.printunlock+0
	rel 75+4 t=8 runtime.morestack_noctxt+0
go.cuinfo.packagename. SDWARFINFO dupok size=0
	0x0000 6d 61 69 6e                                      main
go.string."hello world" SRODATA dupok size=11
	0x0000 68 65 6c 6c 6f 20 77 6f 72 6c 64                 hello world
go.string."hello world\n" SRODATA dupok size=12
	0x0000 68 65 6c 6c 6f 20 77 6f 72 6c 64 0a              hello world.
go.loc."".main SDWARFLOC size=0
go.info."".main SDWARFINFO size=33
	0x0000 03 22 22 2e 6d 61 69 6e 00 00 00 00 00 00 00 00  ."".main........
	0x0010 00 00 00 00 00 00 00 00 00 01 9c 00 00 00 00 01  ................
	0x0020 00                                               .
	rel 9+8 t=1 "".main+0
	rel 17+8 t=1 "".main+81
	rel 27+4 t=30 gofile../root/go/src/github.com/DQNEO/babygo/sample/min.go+0
go.range."".main SDWARFRANGE size=0
go.debuglines."".main SDWARFMISC size=17
	0x0000 04 02 12 0a a5 9c 06 41 06 08 4c 71 04 01 03 7d  .......A..Lq...}
	0x0010 01                                               .
""..inittask SNOPTRDATA size=24
	0x0000 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0010 00 00 00 00 00 00 00 00                          ........
gclocals·33cdeccccebe80329f1fdbee7f5874cb SRODATA dupok size=8
	0x0000 01 00 00 00 00 00 00 00                          ........
gclocals·9fb7f0986f647f17cb53dda1484e0f7a SRODATA dupok size=10
	0x0000 02 00 00 00 01 00 00 00 00 01                    ..........
