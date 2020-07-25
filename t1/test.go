package main

import "syscall"

func writeln(s string) {
	var s2 = s + "\n"
	write(s2)
}

func write(s string) {
	var slc []uint8 = []uint8(s)
	syscall.Write(1, slc)
}


func testInfer() {
	var s = "infer string literal"
	writeln(s)

	var i = 3 + 5
	var j = i
	writeln(itoa(j))
}

func testEscapedChar() {
	var chars []uint8 = []uint8{'\\', '\t', '\r', '\'', '\n'}
	write("start")
	write(string(chars))
	writeln("end")
}

const O_READONLY int = 0

func testOpenRead() {
	var fd int
	fd, _ = syscall.Open("t1/text.txt", O_READONLY, 0)
	writeln(itoa(fd))
	var buf []uint8 = make([]uint8, 300, 300)
	var n int
	n, _ = syscall.Read(fd, buf)
	writeln(itoa(n)) // should be 280
	var readbytes []uint8 = buf[0:n]
	writeln(string(readbytes))
}

func testLogicalAndOr() {
	var t bool = true
	var f bool = false

	if t && t {
		writeln("true && true ok")
	} else {
		writeln("ERROR")
	}
	if t && f {
		writeln("ERROR")
	} else {
		writeln("true && false ok")
	}
	if f && t {
		writeln("ERROR")
	} else {
		writeln("false && true ok")
	}
	if f && f {
		writeln("ERROR")
	} else {
		writeln("false && false ok")
	}

	if t || t {
		writeln("true || true ok")
	} else {
		writeln("ERROR")
	}
	if t || f {
		writeln("true || false ok")
	} else {
		writeln("ERROR")
	}
	if f || t {
		writeln("false || true ok")
	} else {
		writeln("ERROR")
	}
	if f || f {
		writeln("ERROR")
	} else {
		writeln("false || false ok")
	}
}

func testVaargNotPassed(a int, b ...int) {
	if b == nil {
		write(itoa(a))
		writeln(" nil vaargs ok")
	} else {
		writeln("ERROR")
	}
}

func _vaprintf(f string, a ...string) {
	write(Sprintf(f, a))
}

func testVaargs() {
	testVaargNotPassed(777)
	_vaprintf("pass nil slice\n")
	_vaprintf("%s %s %s\n", "a", "bc", "def")
}

const MY_CONST_INT_VALUE int = 24

func testConst() {
	writeln(itoa(MY_CONST_INT_VALUE))
}

func testSwitchString() {
	var testVar string = "foo"
	var caseVar string = "fo"

	switch testVar {
	case "dummy":
		writeln("ERROR")
	}

	switch testVar {
	case "x", caseVar + "o":
		writeln("swithc string 1 ok")
	case "", "y":
		writeln("ERROR")
	default:
		writeln("ERROR")
	}

	switch testVar {
	case "":
		writeln("ERROR")
	case "fo":
		writeln("ERROR")
	default:
		writeln("switch string default ok")
	case "fooo":
		writeln("ERROR")
	}
}

func testSwitchByte() {
	var testVar uint8 = 'c'
	var caseVar uint8 = 'a'
	switch testVar {
	case 'b':
		writeln("ERROR")
	case caseVar + 2:
		writeln("switch uint8 ok")
	default:
		writeln("ERROR")
	}

	switch testVar {
	case 0:
		writeln("ERROR")
	case 'b':
		writeln("ERROR")
	default:
		writeln("switch default ok")
	case 'd':
		writeln("ERROR")
	}
}

func testSwitchInt() {
	var testVar int = 7
	var caseVar int = 5
	switch testVar {
	case 1:
		writeln("ERROR")
	case caseVar + 2:
		writeln("switch int ok")
	default:
		writeln("ERROR")
	}

	switch testVar {
	case 0:
		writeln("ERROR")
	case 6:
		writeln("ERROR")
	default:
		writeln("switch default ok")
	case 8:
		writeln("ERROR")
	}
}

func testStringComparison() {
	var s string
	if s == "" {
		writeln("string cmp 1 ok")
	} else {
		writeln("ERROR")
	}
	var s2 string = ""
	if s2 == s {
		writeln("string cmp 2 ok")
	} else {
		writeln("ERROR")
	}

	var s3 string = "abc"
	s3 = s3 + "def"
	var s4 string = "1abcdef1"
	var s5 string = s4[1:7]
	if s3 == s5 {
		writeln("string cmp 3 ok")
	} else {
		writeln("ERROR")
	}

	if "abcdef" == s5 {
		writeln("string cmp 4 ok")
	} else {
		writeln("ERROR")
	}

	if s3 != s5 {
		writeln(s3)
		writeln(s5)
		writeln("ERROR")
		return
	} else {
		writeln("string cmp not 1 ok")
	}

	if s4 != s3 {
		writeln("string cmp not 2 ok")
	} else {
		writeln("ERROR")
	}
}

func returnTrue1() bool {
	var bol bool
	bol = true
	return bol
}

func returnTrue2() bool {
	var bol bool
	return !bol
}

func returnFalse() bool {
	var bol bool = true
	return !bol
}

func testGlobalBool() {
	globalbool1 = true
	if globalbool1 {
		writeln("globalbool 1 ok")
	} else {
		writeln("ERROR")
	}

	if globalbool2 {
		writeln("ERROR")
	} else {
		writeln("globalbool 2 ok")
	}

	if globalbool3 {
		writeln("ERROR")
	} else {
		writeln("globalbool 3 ok")
	}
}

func testBool() {
	var bol bool = returnTrue1()
	if bol {
		writeln("bool 1 ok")
	} else {
		writeln("ERROR")
	}

	if returnTrue2() {
		writeln("bool 2 ok")
	} else {
		writeln("ERROR")
	}

	if returnFalse() {
		writeln("ERROR")
	} else {
		writeln("bool 3 ok")
	}
}

func testNilComparison() {
	var p *MyStruct
	if p == nil {
		writeln("nil pointer 1 ok")
	} else {
		writeln("ERROR")
	}
	p = nil
	if p == nil {
		writeln("nil pointer 2 ok")
	} else {
		writeln("ERROR")
	}

	var slc []string
	if slc == nil {
		writeln("nil pointer 3 ok")
	} else {
		writeln("ERROR")
	}
	slc = nil
	if slc == nil {
		writeln("nil pointer 4 ok")
	} else {
		writeln("ERROR")
	}
}

func testSliceLiteral() {
	var slc []string = []string{"this is ", "slice literal"}
	writeln(slc[0] + slc[1])
}

func testArrayCopy() {
	var aInt [3]int = [3]int{1, 2, 3}
	var bInt [3]int = aInt
	aInt[1] = 20

	write(itoa(aInt[1]))
	write(itoa(bInt[1]))
	write("\n")
}

func testLocalArrayWithMoreTypes() {
	var aInt [3]int = [3]int{1, 2, 3}
	var i int
	for _, i = range aInt {
		writeln(itoa(i))
	}

	var aString [3]string = [3]string{"a", "bb", "ccc"}
	var s string
	for _, s = range aString {
		write(s)
	}
	write("\n")

	var aByte [4]uint8 = [4]uint8{'x', 'y', 'z', 10}
	var buf []uint8 = aByte[0:4]
	write(string(buf))
}

func testLocalArray() {
	var aInt [3]int = [3]int{1, 2, 3,}
	write(itoa(aInt[0]))
	write(itoa(aInt[1]))
	write(itoa(aInt[2]))
	write("\n")
}

func Sprintf(format string, a []string) string {

	var buf []uint8
	var inPercent bool
	var argIndex int
	var c uint8
	for _, c = range []uint8(format) {
		if inPercent {
			if c == '%' {
				buf = append(buf, c)
			} else {
				var arg string = a[argIndex]
				argIndex++
				var s string = arg // // p.printArg(arg, c)
				var _c uint8
				for _, _c = range []uint8(s) {
					buf = append(buf, _c)
				}
			}
			inPercent = false
		} else {
			if c == '%' {
				inPercent = true
			} else {
				buf = append(buf, c)
			}
		}
	}

	return string(buf)
}

func testSprintf() {
	var a []string = make([]string, 3, 3)
	a[0] = itoa(1234)
	a[1] = "c"
	a[2] = "efg"
	var s string = Sprintf("%sab%sd%s", a)
	write(s)

	var s2 string = Sprintf("%%rax", nil)
	write(s2)
	write("|\n")
}

func testAppendSlice() {
	var slcslc [][]string
	var slc []string
	slc = append(slc, "aa")
	slc = append(slc, "bb")
	slcslc = append(slcslc, slc)
	slcslc = append(slcslc, slc)
	slcslc = append(slcslc, slc)
	var s string
	for _, slc = range slcslc {
		for _, s = range slc {
			write(s)
		}
		write("|")
	}
	write("\n")
}

func testAppendPtr() {
	var slc []*MyStruct
	var p *MyStruct
	var i int
	for i = 0; i < 10; i++ {
		p = new(MyStruct)
		p.field1 = i
		slc = append(slc, p)
	}

	for _, p = range slc {
		write(itoa(p.field1)) // 123456789
	}
	write("\n")
}


func testAppendString() {
	var slc []string
	slc = append(slc, "a")
	slc = append(slc, "bcde")
	var elm string = "fghijklmn\n"
	slc = append(slc, elm)
	var s string
	for _, s = range slc {
		write(s)
	}
	writeln(itoa(len(slc))) // 3
}

func testAppendInt() {
	var slc []int
	slc = append(slc, 1)
	var i int
	for i = 2; i < 10; i++ {
		slc = append(slc, i)
	}
	slc = append(slc, 10)

	for _, i = range slc {
		write(itoa(i)) // 12345678910
	}
	write("\n")
}

func testAppendByte() {
	var slc []uint8
	var char uint8
	for char = 'a'; char <= 'z'; char++ {
		slc = append(slc, char)
	}
	slc = append(slc, 10) // '\n'
	write(string(slc))
	writeln(itoa(len(slc))) // 27
}

func testSringIndex() {
	var s string = "abcde"
	var char uint8 = s[3]
	writeln(itoa(int(char)))
}

func testSubstring() {
	var s string = "abcdefghi"
	var subs string = s[2:5] // "cde"
	writeln(subs)
}

func testSliceOfSlice() {
	var slc []uint8 = make([]uint8, 3, 3)
	slc[0] = 'a'
	slc[1] = 'b'
	slc[2] = 'c'
	writeln(string(slc))

	var slc1 []uint8 = slc[0:3]
	writeln(string(slc1))

	var slc2 []uint8 = slc[0:2]
	writeln(string(slc2))

	var slc3 []uint8 = slc[1:3]
	writeln(string(slc3))
}

func testForrangeKey() {
	var i int
	var slc []string
	var s string
	slc = make([]string, 3, 3)
	slc[0] = "a"
	slc[1] = "b"
	slc[2] = "c"
	for i, s = range slc {
		write(itoa(i))
		writeln(s)
	}
}

func testForrange() {
	var slc []string
	var s string

	writeln("going to loop 0 times")
	for _, s = range slc {
		write(s)
		write("ERROR")
	}

	slc = make([]string, 2, 2)
	slc[0] = ""
	slc[1] = ""

	writeln("going to loop 2 times")
	for _, s = range slc {
		write(s)
		writeln(" in loop")
	}

	writeln("going to loop 4 times")
	var a int
	for _, a = range globalintarray {
		write(itoa(a))
	}
	writeln("")

	slc = make([]string, 3, 3)
	slc[0] = "hello"
	slc[1] = "for"
	slc[2] = "range"
	for _, s = range slc {
		write(s)
	}
	writeln("")
}

func newStruct() *MyStruct {
	var strct = new(MyStruct)
	writeln(itoa(strct.field2))
	strct.field2 = 2
	return strct
}

func testNewStruct() {
	var strct *MyStruct
	strct = newStruct()
	writeln(itoa(strct.field1))
	writeln(itoa(strct.field2))
}

var nilSlice []*MyStruct

func testNilSlice() {
	writeln("-- testNilSlice()")
	nilSlice = make([]*MyStruct, 2, 2)
	writeln(itoa(len(nilSlice)))
	writeln(itoa(cap(nilSlice)))

	nilSlice = nil
	writeln(itoa(len(nilSlice)))
	writeln(itoa(cap(nilSlice)))
}

func testZeroValues() {
	writeln("-- testZeroValues()")
	var s string
	write(s)

	var s2 string = ""
	write(s2)
	var h int = 1
	var i int
	var j int = 2
	writeln(itoa(h))
	writeln(itoa(i))
	writeln(itoa(j))

	if i == 0 {
		writeln("int zero ok")
	} else {
		writeln("ERROR")
	}
}

func testIncrDecr() {
	var i int = 0
	i++
	writeln(itoa(i))

	i--
	i--
	writeln(itoa(i))
}

type MyStruct struct {
	field1 int
	field2 int
}

var globalstrings1 [2]string
var globalstrings2 [2]string
var __slice []string

func testGlobalStrings() {
	globalstrings1[0] = "aaa,"
	globalstrings1[1] = "bbb,"
	globalstrings2[0] = "ccc,"
	globalstrings2[1] = "ddd,"
	__slice = make([]string, 1, 1)
	write(globalstrings1[0])
	write(globalstrings1[1])
	write(globalstrings1[0])
	write(globalstrings1[1])
}

var globalstrings [2]string

func testSliceOfStrings() {
	var s1 string = "hello"
	var s2 string = " strings\n"
	var strings []string = make([]string, 2, 2)
	var i int
	strings[0] = s1
	strings[1] = s2
	for i = 0; i < 2; i = i + 1 {
		write(strings[i])
	}

	globalstrings[0] = s1
	globalstrings[1] = " globalstrings\n"
	for i = 0; i < 2; i = i + 1 {
		write(globalstrings[i])
	}
}

var structPointers []*MyStruct

func testSliceOfPointers() {
	var strct1 MyStruct
	var strct2 MyStruct
	var p1 *MyStruct = &strct1
	var p2 *MyStruct = &strct2

	strct1.field2 = 11
	strct2.field2 = 22
	structPointers = make([]*MyStruct, 2, 2)
	structPointers[0] = p1
	structPointers[1] = p2

	var i int
	for i = 0; i < 2; i = i + 1 {
		writeln(itoa(structPointers[i].field2))
	}
}

func testStructPointer() {
	var _strct MyStruct
	var strct *MyStruct
	strct = &_strct
	strct.field1 = 123

	var i int
	i = strct.field1
	writeln(itoa(i))

	strct.field2 = 456
	writeln(itoa(_strct.field2))

	strct.field1 = 777
	strct.field2 = strct.field1
	writeln(itoa(strct.field2))

}

func testStruct() {
	var strct MyStruct
	strct.field1 = 123

	var i int
	i = strct.field1
	writeln(itoa(i))

	strct.field2 = 456
	writeln(itoa(strct.field2))

	strct.field1 = 777
	strct.field2 = strct.field1
	writeln(itoa(strct.field2))
}

func testPointer() {
	var i int = 12
	var j int
	var p *int
	p = &i
	j = *p
	writeln(itoa(j))
	*p = 11
	writeln(itoa(i))

	var c uint8 = 'A'
	var pc *uint8
	pc = &c
	*pc = 'B'
	var slc []uint8
	slc = make([]uint8, 1, 1)
	slc[0] = c
	writeln(string(slc))
}

func testDeclValue() {
	var i int = 123
	writeln(itoa(i))
}

func testConcateStrings() {
	var concatenated string = "foo" + "bar" + "1234"
	writeln(concatenated)
}

func testLenCap() {
	var x []uint8
	x = make([]uint8, 0, 0)
	writeln(itoa(len(x)))

	writeln(itoa(cap(x)))

	x = make([]uint8, 12, 24)
	writeln(itoa(len(x)))

	writeln(itoa(cap(x)))

	writeln(itoa(len(globalintarray)))

	writeln(itoa(cap(globalintarray)))

	var s string
	s = "hello\n"
	writeln(itoa(len(s))) // 6
}

func testMakeSlice() {
	var x []uint8 = make([]uint8, 3, 20)
	x[0] = 'A'
	x[1] = 'B'
	x[2] = 'C'
	writeln(string(x))
}

func testNew() {
	var i *int
	i = new(int)
	writeln(itoa(*i)) // 0
}

func testItoa() {
	writeln(itoa(1234567890))
	writeln(itoa(54321))
	writeln(itoa(1))
	writeln("0")
	writeln(itoa(0))
	writeln(itoa(-1))
	writeln(itoa(-54321))
	writeln(itoa(-1234567890))
}

func itoa(ival int) string {
	var buf []uint8 = make([]uint8, 100,100)
	var r []uint8 = make([]uint8, 100, 100)

	var next int
	var right int
	var ix int = 0
	if ival == 0 {
		return "0"
	}
	var minus bool
	minus = false
	for ix = 0; ival != 0; ix = ix + 1 {
		if ival < 0 {
			ival = -1 * ival
			minus = true
			r[0] = '-'
		} else {
			next = ival / 10
			right = ival - next*10
			ival = next
			buf[ix] = uint8('0' + right)
		}
	}

	var j int
	var c uint8
	for j = 0; j < ix; j = j + 1 {
		c = buf[ix-j-1]
		if minus {
			r[j+1] = c
		} else {
			r[j] = c
		}
	}

	return string(r[0:ix])
}

func testForOmissible() {
	var i int
	for {
		i++
		if i == 2 {
			break
		}
	}
	write(itoa(i))

	i = 0
	for i < 3 {
		i++
	}
	write(itoa(i))

	i = 0
	for i < 4 {
		i++
	}
	write(itoa(i))

	write("\n")
}

func testForBreakContinue() {
	var i int
	for i = 0; i < 10; i = i + 1 {
		if i == 3 {
			break
		}
		write(itoa(i))
	}
	write("exit")
	writeln(itoa(i))
	for i = 0; i < 10; i = i + 1 {
		if i < 7 {
			continue
		}
		write(itoa(i))
	}
	write("exit")
	writeln(itoa(i))

	var ary []int = []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	for _, i = range ary {
		if i == 3 {
			break
		}
		write(itoa(i))
	}
	write("exit")
	writeln(itoa(i))
	for _, i = range ary {
		if i < 7 {
			continue
		}
		write(itoa(i))
	}
	write("exit")
	writeln(itoa(i))
}

func testFor() {
	var i int
	for i = 0; i < 3; i = i + 1 {
		write("A")
	}
	write("\n")
}

func testCmpUint8() {
	var localuint8 uint8 = 1
	if localuint8 == 1 {
		writeln("uint8 cmp == ok")
	}
	if localuint8 != 1 {
		writeln("ERROR")
	} else {
		writeln("uint8 cmp != ok")
	}
	if localuint8 > 0 {
		writeln("uint8 cmp > ok")
	}
	if localuint8 < 0 {
		writeln("ERROR")
	} else {
		writeln("uint8 cmp < ok")
	}

	if localuint8 >= 1 {
		writeln("uint8 cmp >= ok")
	}
	if localuint8 <= 1 {
		writeln("uint8 cmp <= ok")
	}

	localuint8 = 101
	if localuint8 == 'A' {
		writeln("uint8 cmp == A ok")
	}
}

func testCmpInt() {
	var a int = 1
	if a == 1 {
		writeln("int cmp == ok")
	}
	if a != 1 {
		writeln("ERROR")
	} else {
		writeln("int cmp != ok")
	}
	if a > 0 {
		writeln("int cmp > ok")
	}
	if a < 0 {
		writeln("ERROR")
	} else {
		writeln("int cmp < ok")
	}

	if a >= 1 {
		writeln("int cmp >= ok")
	}
	if a <= 1 {
		writeln("int cmp <= ok")
	}
	a = 101
	if a == 'A' {
		writeln("int cmp == A ok")
	}
}

func testIf() {
	var tr bool = true
	var fls bool = false
	if tr {
		writeln("ok true")
	}
	if fls {
		writeln("ERROR")
	}
	writeln("ok false")
}

func testElse() {
	if true {
		writeln("ok true")
	} else {
		writeln("ERROR")
	}

	if false {
		writeln("ERROR")
	} else {
		writeln("ok false")
	}
}

var globalbool1 bool // = true // @TODO
var globalbool2 bool = false
var globalbool3 bool
var globalint int
var globalint2 int
var globaluint8 uint8
var globaluint16 uint16

var globalstring string
var globalstring2 string
var globalintslice []int
var globalarray [10]uint8
var globalslice []uint8
var globaluintptr uintptr

func assignGlobal() {
	globalint = 22
	globaluint8 = 1
	globaluint16 = 5
	globaluintptr = 7
	globalstring = "globalstring changed\n"
}

func add1(x int) int {
	return x + 1
}

func sum(x int, y int) int {
	return x + y
}

func print1(a string) {
	write(a)
	return
}

func print2(a string, b string) {
	write(a)
	write(b)
}

func returnstring() string {
	return "i am a local 1\n"
}

var globalintarray [4]int

func testIndexExprOfArray() {
	globalintarray[0] = 11
	globalintarray[1] = 22
	globalintarray[2] = globalintarray[1]
	globalintarray[3] = 44
	write("\n")
}

func testIndexExprOfSlice() {
	var intslice []int = globalintarray[0:4]
	intslice[0] = 66
	intslice[1] = 77
	intslice[2] = intslice[1]
	intslice[3] = 88

	var i int
	for i = 0; i < 4; i = i + 1 {
		write(itoa(intslice[i]))
	}
	write("\n")

	for i = 0; i < 4; i = i + 1 {
		write(itoa(globalintarray[i]))
	}
	write("\n")
}

// test global chars
func testChar() {
	globalarray[0] = 'A'
	globalarray[1] = 'B'
	globalarray[2] = globalarray[0]
	globalarray[3] = 100 / 10 // '\n'
	globalarray[1] = 'B'
	var chars []uint8 = globalarray[0:4]
	write(string(chars))
	globalslice = chars
	write(string(globalarray[0:4]))
}

func testString() {
	write(globalstring)
	assignGlobal()

	print1("hello string literal\n")

	var localstring1 string = returnstring()
	var localstring2 string
	localstring2 = "i m local2\n"
	print2(localstring1, localstring2)
	write(globalstring)
}

func testArgAssign(x int) int {
	x = 13
	return x
}

func testMinus() int {
	var x int = -1
	x = x * -5
	return x
}

func testMisc() {
	var i13 int = 0
	i13 = testArgAssign(i13)
	var i5 int = testMinus()
	globalint2 = sum(1, i13%i5)

	var locali3 int
	var tmp int
	tmp = int(uint8('3' - '1'))
	tmp = tmp + int(globaluint16)
	tmp = tmp + int(globaluint8)
	tmp = tmp + int(globaluintptr)
	locali3 = add1(tmp)
	var i42 int
	i42 = sum(globalint, globalint2) + locali3

	writeln(itoa(i42))
}

var globalptr *int

func test() {
	testInfer()
	testEscapedChar()
	testOpenRead()
	testVaargs()
	testSwitchString()
	testSwitchByte()
	testSwitchInt()
	testLogicalAndOr()
	testConst()
	testForOmissible()
	testForBreakContinue()
	testBool()
	testGlobalBool()
	testNilComparison()
	testSliceLiteral()
	testArrayCopy()
	testLocalArrayWithMoreTypes()
	testLocalArray()
	testSprintf()
	testAppendSlice()
	testAppendPtr()
	testAppendString()
	testAppendInt()
	testAppendByte()
	testSringIndex()
	testSubstring()
	testSliceOfSlice()
	testForrangeKey()
	testForrange()
	testNewStruct()
	testNilSlice()
	testZeroValues()
	testIncrDecr()
	testGlobalStrings()
	testSliceOfStrings()
	testSliceOfStrings()
	testSliceOfPointers()
	testStructPointer()
	testStruct()
	testPointer()
	testDeclValue()
	testStringComparison()
	testConcateStrings()
	testLenCap()
	testMakeSlice()
	testNew()

	testItoa()
	testIndexExprOfArray()
	testIndexExprOfSlice()
	testString()
	testFor()
	testCmpUint8()
	testCmpInt()
	testIf()
	testElse()
	testChar()
	testMisc()
	print("test end\n")
}

func main() {
	test()
}
