package main

import "os"

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

func assign() {
	globalint = 30
	globaluint8 = 1
	globaluint16 = 16
	globaluintptr = 20
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
	globalarray[3] = 10 // '\n'
	printchars()
}

func main() {
	assign()
	testChar()
	globalint2 = sum(1, 2)

	print(globalstring)

	var localstring1 string
	var localstring2 string

	print1("hello string literal\n")
	localstring1 = returnstring()
	localstring2 = "i m local2\n"
	print2(localstring1, localstring2)
	globalstring = "globalstring changed\n"
	print(globalstring)
	var locali3 int
	var eight int
	eight = '3' - '1'
	eight = eight + 6
	locali3 = add1(eight)
	os.Exit( sum(globalint , globalint2) + locali3)
}
