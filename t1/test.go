package main

import (
	"syscall"
)

func testAppendPtr() {
	var slc []*MyStruct
	var p *MyStruct
	var i int
	for i=0;i<10;i++ {
		p = new(MyStruct)
		p.field1 = i
		slc = append(slc, p)
	}

	for _, p = range slc {
		write(Itoa(p.field1)) // 123456789
	}
	write("\n")
}


func testAppendInt() {
	var slc []int
	slc = append(slc, 1)
	var i int
	for i=2;i<10;i++ {
		slc = append(slc, i)
	}
	slc = append(slc, 10)

	for _, i = range slc {
		write(Itoa(i)) // 12345678910
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
	writeln(Itoa(len(slc))) // 3
}

func testAppendByte() {
	var slc []uint8
	var char uint8
	for char = 'a'; char <= 'z' ; char++ {
		slc = append(slc, char)
	}
	slc = append(slc, 10) // '\n'
	write(string(slc))
	writeln(Itoa(len(slc))) // 27
}

func testSliceOfSlice() {
	var slc []uint8 = make([]uint8, 3,3)
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
		write(Itoa(a))
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

func _testNew() *MyStruct {
	var strct *MyStruct
	strct = new(MyStruct)
	writeln(Itoa(strct.field2))
	strct.field2 = 2
	return strct
}

func testNew() {
	var strct *MyStruct
	strct = _testNew()
	writeln(Itoa(strct.field1))
	writeln(Itoa(strct.field2))
	var i *int
	i = new(int)
	writeln(Itoa(*i)) // 0
}

var testNilSlice []*MyStruct

func testNil() {
	writeln("-- testNil()")
	testNilSlice = make([]*MyStruct, 2, 2)
	writeln(Itoa(len(testNilSlice)))
	writeln(Itoa(cap(testNilSlice)))

	testNilSlice = nil
	writeln(Itoa(len(testNilSlice)))
	writeln(Itoa(cap(testNilSlice)))
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
	writeln(Itoa(h))
	writeln(Itoa(i))
	writeln(Itoa(j))

	if i == 0 {
		writeln("int zero ok")
	} else {
		writeln("ERROR")
	}
}

func testIncrDecr() {
	var i int = 0
	i++
	writeln(Itoa(i))

	i--
	i--
	writeln(Itoa(i))
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
	var strings []string = make([]string, 2,2)
	var i int
	strings[0] = s1
	strings[1] = s2
	for i = 0; i<2;i = i+1 {
		write(strings[i])
	}

	globalstrings[0] = s1
	globalstrings[1] = " globalstrings\n"
	for i = 0; i<2;i = i+1 {
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
	for i = 0; i<2;i = i+1 {
		writeln(Itoa(structPointers[i].field2))
	}
}

func testStructPointer() {
	var _strct MyStruct
	var strct *MyStruct
	strct = &_strct
	strct.field1 = 123

	var i int
	i  = strct.field1
	writeln(Itoa(i))

	strct.field2 = 456
	writeln(Itoa(_strct.field2))

	strct.field1 = 777
	strct.field2 = strct.field1
	writeln(Itoa(strct.field2))

}

func testStruct() {
	var strct MyStruct
	strct.field1 = 123

	var i int
	i  = strct.field1
	writeln(Itoa(i))

	strct.field2 = 456
	writeln(Itoa(strct.field2))

	strct.field1 = 777
	strct.field2 = strct.field1
	writeln(Itoa(strct.field2))
}

func testPointer() {
	var i int = 12
	var j int
	var p *int
	p = &i
	j = *p
	writeln(Itoa(j))
	*p = 11
	writeln(Itoa(i))

	var c uint8 = 'A'
	var pc *uint8
	pc = &c
	*pc = 'B'
	var slc []uint8
	slc = make([]uint8, 1 , 1)
	slc[0] = c
	writeln(string(slc))
}

func testDeclValue() {
	var i int = 123
	writeln(Itoa(i))
}

func testConcateStrings() {
	var concatenated string = "foo" + "bar" + "1234"
	writeln(concatenated)
}

func testLen() {
	var x []uint8
	x = make([]uint8, 0 , 0)
	writeln(Itoa(len(x)))

	writeln(Itoa(cap(x)))

	x = make([]uint8, 12, 24)
	writeln(Itoa(len(x)))

	writeln(Itoa(cap(x)))

	writeln(Itoa(len(globalintarray)))

	writeln(Itoa(cap(globalintarray)))

	var s string
	s = "hello\n"
	writeln(Itoa(len(s))) // 6
}

func testMalloc() {
	var x []uint8 = make([]uint8, 3, 20)
	x[0] = 'A'
	x[1] = 'B'
	x[2] = 'C'
	writeln(string(x))
}

func testMakaSlice() []uint8 {
	var slc []uint8 = make([]uint8, 0, 10)
	return slc
}

func testItoa() {
	writeln(Itoa(1234567890))
	writeln(Itoa(54321))
	writeln(Itoa(1))
	writeln("0")
	writeln(Itoa(0))
	writeln(Itoa(-1))
	writeln(Itoa(-54321))
	writeln(Itoa(-1234567890))
}

var buf [100]uint8
var r [100]uint8

func Itoa(ival int) string {
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

func testFor() {
	var i int
	for i=0;i<3; i = i + 1 {
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

var globalint int = 30
var globalint2 int = 0
var globaluint8 uint8 = 8
var globaluint16 uint16 = 16

var globalstring string = "hello stderr\n"
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

// test global chars
func testChar() {
	globalarray[0] = 'A'
	globalarray[1] = 'B'
	globalarray[2] = globalarray[0]
	globalarray[3] = 100 / 10 // '\n'

	var chars []uint8 = globalarray[0:4]
	write(string(chars))
	globalslice = chars
	write(string(globalarray[0:4]))
}

var globalintarray [4]int

func testIndexExprOfArray() {
	globalintarray[0] = 11
	globalintarray[1] = 22
	globalintarray[2] = globalintarray[1]
	globalintarray[3] = 44
	/*
		var i int
	for i = 0; i<4 ;i= i+1 {
		//write("x")
		Itoa(globalintarray[i])
	}

	 */
	write("\n")
}

func testIndexExprOfSlice() {
	var intslice []int = globalintarray[0:4]
	intslice[0] = 66
	intslice[1] = 77
	intslice[2] = intslice[1]
	intslice[3] = 88

	var i int
	for i = 0; i<4 ;i= i+1 {
		write(Itoa(intslice[i]))
	}
	write("\n")

	for i = 0; i<4 ;i= i+1 {
		write(Itoa(globalintarray[i]))
	}
	write("\n")
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

func testMisc() {
	var i13 int = 0
	i13 = testArgAssign(i13)
	var i5 int = testMinus()
	globalint2 = sum(1, i13 % i5)

	var locali3 int
	var tmp int
	tmp = int(uint8('3' - '1'))
	tmp = tmp + int(globaluint16)
	tmp = tmp + int(globaluint8)
	tmp = tmp + int(globaluintptr)
	locali3 = add1(tmp)
	var i42 int
	i42 =  sum(globalint , globalint2) + locali3

	writeln(Itoa(i42))
}

func writeln(s string) {
	var s2 string = s + "\n"
	write(s2)
}
func write(s string) {
	var slc []uint8 = []uint8(s)
	syscall.Write(1, slc)
}

var globalptr *int

func test() {
	testAppendPtr()
	testAppendInt()
	testAppendString()
	testAppendByte()
	testSliceOfSlice()
	testForrange()
	testNew()
	testNil()
	testZeroValues()
	testIncrDecr()
	testSliceOfStrings()
	testSliceOfStrings()
	testSliceOfPointers()
	testStructPointer()
	testStruct()
	testPointer()
	testDeclValue()
	testConcateStrings()
	testLen()
	testMalloc()
	testIndexExprOfArray()
	testIndexExprOfSlice()
	testItoa()
	testFor()
	testCmpUint8()
	testCmpInt()
	testIf()
	testElse()
	testChar()
	testString()
	testMisc()
}

func main() {
	test()
}
