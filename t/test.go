package main

import (
	"syscall"
)

func testPointer() {
	var i int = 12
	var j int
	var p *int
	p = &i
	j = *p
	write(Itoa(j))
	write("\n")
	*p = 11
	write(Itoa(i))
	write("\n")
}

func testDeclValue() {
	var i int = 123
	write(Itoa(i))
	write("\n")
}

func testConcateStrings() {
	var concatenated string = "foo" + "bar" + "1234\n"
	write(concatenated)
}

func testLen() {
	var x []uint8
	x = make([]uint8, 0 , 0)
	write(Itoa(len(x)))
	write("\n")

	write(Itoa(cap(x)))
	write("\n")

	x = make([]uint8, 12, 24)
	write(Itoa(len(x)))
	write("\n")

	write(Itoa(cap(x)))
	write("\n")

	write(Itoa(len(globalintarray)))
	write("\n")

	write(Itoa(cap(globalintarray)))
	write("\n")

	var s string
	s = "hello\n"
	write(Itoa(len(s))) // 6
	write("\n")
}

func testMalloc() {
	var x []uint8 = make([]uint8, 3, 20)
	x[0] = 'A'
	x[1] = 'B'
	x[2] = 'C'
	write(string(x))
	write("\n")
}

func testMakaSlice() []uint8 {
	var slc []uint8 = make([]uint8, 0, 10)
	return slc
}

func testItoa() {
	write(Itoa(1234567890))
	write("\n")
	write(Itoa(54321))
	write("\n")
	write(Itoa(1))
	write("\n")
	write("0")
	write("\n")
	write(Itoa(0))
	write("\n")
	write(Itoa(-1))
	write("\n")
	write(Itoa(-54321))
	write("\n")
	write(Itoa(-1234567890))
	write("\n")
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
		write("uint8 cmp == ok\n")
	}
	if localuint8 != 1 {
		write("ERROR\n")
	} else {
		write("uint8 cmp != ok\n")
	}
	if localuint8 > 0 {
		write("uint8 cmp > ok\n")
	}
	if localuint8 < 0 {
		write("ERROR\n")
	} else {
		write("uint8 cmp < ok\n")
	}

	if localuint8 >= 1 {
		write("uint8 cmp >= ok\n")
	}
	if localuint8 <= 1 {
		write("uint8 cmp <= ok\n")
	}
}

func testCmpInt() {
	var a int = 1
	if a == 1 {
		write("int cmp == ok\n")
	}
	if a != 1 {
		write("ERROR\n")
	} else {
		write("int cmp != ok\n")
	}
	if a > 0 {
		write("int cmp > ok\n")
	}
	if a < 0 {
		write("ERROR\n")
	} else {
		write("int cmp < ok\n")
	}

	if a >= 1 {
		write("int cmp >= ok\n")
	}
	if a <= 1 {
		write("int cmp <= ok\n")
	}

}

func testIf() {
	var tr bool = true
	var fls bool = false

	if tr {
		write("ok true\n")
	}
	if fls {
		write("ERROR\n")
	}
	write("ok false\n")
}

func testElse() {
	if true {
		write("ok true\n")
	} else {
		write("ERROR\n")
	}

	if false {
		write("ERROR\n")
	} else {
		write("ok false\n")
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

	write(Itoa(i42))
	write("\n")
}

func write(s string) {
	var slc []uint8 = []uint8(s)
	syscall.Write(1, slc)
}

var globalptr *int

func test() {
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
