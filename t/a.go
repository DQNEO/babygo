package main

import "os"

func testFor() {
	var i int
	for i=0;i<3; i = i + 1 {
		print("A")
	}
	print("\n")
}

func testCmpUint8() {
	var localuint8 uint8
	localuint8 = 1
	if localuint8 == 1 {
		print("uint8 cmp == ok\n")
	}
	if localuint8 != 1 {
		print("ERROR\n")
	} else {
		print("uint8 cmp != ok\n")
	}
	if localuint8 > 0 {
		print("uint8 cmp > ok\n")
	}
	if localuint8 < 0 {
		print("ERROR\n")
	} else {
		print("uint8 cmp < ok\n")
	}

	if localuint8 >= 1 {
		print("uint8 cmp >= ok\n")
	}
	if localuint8 <= 1 {
		print("uint8 cmp <= ok\n")
	}
}

func testCmpInt() {
	var a int
	a = 1
	if a == 1 {
		print("int cmp == ok\n")
	}
	if a != 1 {
		print("ERROR\n")
	} else {
		print("int cmp != ok\n")
	}
	if a > 0 {
		print("int cmp > ok\n")
	}
	if a < 0 {
		print("ERROR\n")
	} else {
		print("int cmp < ok\n")
	}

	if a >= 1 {
		print("int cmp >= ok\n")
	}
	if a <= 1 {
		print("int cmp <= ok\n")
	}

}

func testIf() {
	var tr bool
	tr = true
	var fls bool
	fls = false

	if tr {
		print("ok true\n")
	}
	if fls {
		print("ERROR\n")
	}
	print("ok false\n")
}

func testElse() {
	if true {
		print("ok true\n")
	} else {
		print("ERROR\n")
	}

	if false {
		print("ERROR\n")
	} else {
		print("ok false\n")
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
	print(a)
	return
}

func print2(a string, b string) {
	print(a)
	print(b)
}

func returnstring() string {
	return "i am a local 1\n"
}


func printchars() {
	var chars []uint8
	chars = globalarray[0:4]
	print(string(chars))
	globalslice = chars
	print(string(chars))
}

// test global chars
func testChar() {
	globalarray[0] = 'A'
	globalarray[1] = 'B'
	globalarray[2] = globalarray[0]
	globalarray[3] = 100 / 10 // '\n'
	printchars()
}

func testArgAssign(x int) int {
	x = 13
	return x
}

func main() {
	testFor()
	testCmpUint8()
	testCmpInt()
	testIf()
	testElse()
	testChar()
	var i13 int = 0
	i13 = testArgAssign(i13)
	globalint2 = sum(1, i13 % 5)
	print(globalstring)

	assignGlobal()

	var localstring1 string
	var localstring2 string

	print1("hello string literal\n")
	localstring1 = returnstring()
	localstring2 = "i m local2\n"
	print2(localstring1, localstring2)
	print(globalstring)
	var locali3 int
	var tmp int
	tmp = '3' - '1'
	tmp = tmp + int(globaluint16)
	tmp = tmp + int(globaluint8)
	tmp = tmp + int(globaluintptr)
	locali3 = add1(tmp)
	os.Exit( sum(globalint , globalint2) + locali3)
}
