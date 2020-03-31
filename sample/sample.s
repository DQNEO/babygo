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
"".arg1ret1 STEXT nosplit size=19 args=0x10 locals=0x0
	0x0000 00000 (sample/sample.go:13)	TEXT	"".arg1ret1(SB), NOSPLIT|ABIInternal, $0-16
	0x0000 00000 (sample/sample.go:13)	PCDATA	$0, $-2
	0x0000 00000 (sample/sample.go:13)	PCDATA	$1, $-2
	0x0000 00000 (sample/sample.go:13)	FUNCDATA	$0, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0000 00000 (sample/sample.go:13)	FUNCDATA	$1, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0000 00000 (sample/sample.go:13)	FUNCDATA	$2, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x0000 00000 (sample/sample.go:13)	PCDATA	$0, $0
	0x0000 00000 (sample/sample.go:13)	PCDATA	$1, $0
	0x0000 00000 (sample/sample.go:13)	MOVQ	$0, "".~r1+16(SP)
	0x0009 00009 (sample/sample.go:14)	MOVQ	$7, "".~r1+16(SP)
	0x0012 00018 (sample/sample.go:14)	RET
	0x0000 48 c7 44 24 10 00 00 00 00 48 c7 44 24 10 07 00  H.D$.....H.D$...
	0x0010 00 00 c3                                         ...
"".main STEXT nosplit size=64 args=0x0 locals=0x20
	0x0000 00000 (sample/sample.go:17)	TEXT	"".main(SB), NOSPLIT|ABIInternal, $32-0
	0x0000 00000 (sample/sample.go:17)	SUBQ	$32, SP
	0x0004 00004 (sample/sample.go:17)	MOVQ	BP, 24(SP)
	0x0009 00009 (sample/sample.go:17)	LEAQ	24(SP), BP
	0x000e 00014 (sample/sample.go:17)	PCDATA	$0, $-2
	0x000e 00014 (sample/sample.go:17)	PCDATA	$1, $-2
	0x000e 00014 (sample/sample.go:17)	FUNCDATA	$0, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x000e 00014 (sample/sample.go:17)	FUNCDATA	$1, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x000e 00014 (sample/sample.go:17)	FUNCDATA	$2, gclocals·33cdeccccebe80329f1fdbee7f5874cb(SB)
	0x000e 00014 (sample/sample.go:18)	PCDATA	$0, $-1
	0x000e 00014 (sample/sample.go:18)	PCDATA	$1, $-1
	0x000e 00014 (sample/sample.go:18)	JMP	16
	0x0010 00016 (sample/sample.go:19)	PCDATA	$0, $0
	0x0010 00016 (sample/sample.go:19)	PCDATA	$1, $0
	0x0010 00016 (sample/sample.go:19)	MOVQ	$1, "".x+16(SP)
	0x0019 00025 (sample/sample.go:19)	JMP	27
	0x001b 00027 (sample/sample.go:20)	MOVQ	$2, "".x+8(SP)
	0x0024 00036 (sample/sample.go:20)	MOVQ	$0, "".~r1(SP)
	0x002c 00044 (sample/sample.go:20)	MOVQ	$7, "".~r1(SP)
	0x0034 00052 (sample/sample.go:20)	JMP	54
	0x0036 00054 (sample/sample.go:20)	PCDATA	$0, $-1
	0x0036 00054 (sample/sample.go:20)	PCDATA	$1, $-1
	0x0036 00054 (sample/sample.go:20)	MOVQ	24(SP), BP
	0x003b 00059 (sample/sample.go:20)	ADDQ	$32, SP
	0x003f 00063 (sample/sample.go:20)	RET
	0x0000 48 83 ec 20 48 89 6c 24 18 48 8d 6c 24 18 eb 00  H.. H.l$.H.l$...
	0x0010 48 c7 44 24 10 01 00 00 00 eb 00 48 c7 44 24 08  H.D$.......H.D$.
	0x0020 02 00 00 00 48 c7 04 24 00 00 00 00 48 c7 04 24  ....H..$....H..$
	0x0030 07 00 00 00 eb 00 48 8b 6c 24 18 48 83 c4 20 c3  ......H.l$.H.. .
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
	rel 13+8 t=1 "".arg1ret1+19
	rel 24+4 t=29 go.info."".arg1ret1$abstract+13
	rel 37+4 t=29 go.info.int+0
go.range."".arg1ret1 SDWARFRANGE size=0
go.debuglines."".arg1ret1 SDWARFMISC size=14
	0x0000 04 02 03 07 14 6a 06 69 04 01 03 73 06 01        .....j.i...s..
go.loc."".main SDWARFLOC size=0
go.info."".main SDWARFINFO size=33
	0x0000 03 22 22 2e 6d 61 69 6e 00 00 00 00 00 00 00 00  ."".main........
	0x0010 00 00 00 00 00 00 00 00 00 01 9c 00 00 00 00 01  ................
	0x0020 00                                               .
	rel 0+0 t=24 type.int+0
	rel 9+8 t=1 "".main+0
	rel 17+8 t=1 "".main+64
	rel 27+4 t=30 gofile../mnt/sample/sample.go+0
go.range."".main SDWARFRANGE size=0
go.debuglines."".main SDWARFMISC size=21
	0x0000 04 02 0a 03 0b 14 9c 24 06 69 06 24 06 69 06 c3  .......$.i.$.i..
	0x0010 04 01 03 6d 01                                   ...m.
""..inittask SNOPTRDATA size=24
	0x0000 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00 00  ................
	0x0010 00 00 00 00 00 00 00 00                          ........
gclocals·33cdeccccebe80329f1fdbee7f5874cb SRODATA dupok size=8
	0x0000 01 00 00 00 00 00 00 00                          ........
